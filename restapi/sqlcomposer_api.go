package restapi

import (
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	log "github.com/sirupsen/logrus"
	"github.com/user/sqlcomposer-svc/models"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"github.com/wangxb07/sqlcomposer"
	"gopkg.in/yaml.v2"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
)

type SqlComposerRequest struct {
	PageIndex int64                    `json:"page_index"`
	PageLimit int64                    `json:"page_limit"`
	Filters   []*SqlComposerFilterItem `json:"filters"`
	Sorts     [][]string               `json:"sorts"`
}

type SqlComposerFilterItem struct {
	Attr string               `json:"attr"`
	Op   sqlcomposer.Operator `json:"op"`
	Val  interface{}          `json:"val"`
}

var db *sqlx.DB

type Config struct {
	DB *sqlx.DB
}

func Setup(cfg *Config) {
	//init db
	db = cfg.DB
}

func errJSON(err error) map[string]interface{} {
	return map[string]interface{}{
		"err": err.Error(),
	}
}

func errJSONWithSQL(err error, s string) map[string]interface{} {
	return map[string]interface{}{
		"err": err.Error(),
		"sql": s,
	}
}

// @Summary 获取查询结果
// @Tags 接口
// @version 1.0
// @Param path path string true "path"
// @Param debug query string true "debug"
// @Success 200 {string} string	"json"
// @Failure 400 {object} Error "error"
// @Failure 404 {object} Error "not found"
// @Router /{path} [get]
func SqlComposerHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		//get yml by path from db
		path := c.Param("path")
		debug := c.Query("debug")

		docFound, err := models.Docs(qm.Where("path = ?", path)).One(c, db)

		if err != nil {
			log.Error(err)
			c.JSON(http.StatusNotFound, errJSON(err))
			return
		}

		var req SqlComposerRequest
		if err := c.BindJSON(&req); err != nil {
			log.Error(err)
			c.JSON(http.StatusBadRequest, errJSON(err))
			return
		}

		var custFilters []sqlcomposer.Filter
		for _, filter := range req.Filters {
			custFilter := sqlcomposer.Filter{
				Val:  filter.Val,
				Op:   filter.Op,
				Attr: filter.Attr,
			}
			custFilters = append(custFilters, custFilter)
		}

		sorts := &sqlcomposer.OrderBy{}
		for _, s := range req.Sorts {
			var d sqlcomposer.Direction

			d = sqlcomposer.DESC

			if strings.ToUpper(s[1]) == "ASC" {
				d = sqlcomposer.ASC
			}

			*sorts = append(*sorts, sqlcomposer.Sort{
				Name:      s[0],
				Direction: d,
			})
		}

		var doc sqlcomposer.SqlApiDoc

		buffer := []byte(docFound.Content.String)
		err = yaml.Unmarshal(buffer, &doc)
		if err != nil {
			log.Error(err)
			c.JSON(http.StatusBadRequest, errJSON(err))
		}

		dbc, err := models.DatabaseConfigs(qm.Where("name = ?", docFound.DBName)).One(c, db)
		if err != nil {
			log.Error(err)
			c.JSON(http.StatusBadRequest, errJSON(err))
			return
		}

		db, err := sqlx.Connect("mysql", dbc.DSN.String)
		if err != nil {
			log.WithField("dsn", dbc.DSN.String).Error(err)
			c.JSON(http.StatusBadRequest, errJSON(err))
			return
		}

		defer db.Close()

		sqlBuilder, err := sqlcomposer.NewSqlBuilder(db, []byte(docFound.Content.String))
		if err != nil {
			log.Error(err)
			c.JSON(http.StatusBadRequest, errJSON(err))
			return
		}

		configureSqlCompose(sqlBuilder)

		result := struct {
			Total    int64             `json:"total,omitempty"`
			Data     []interface{}     `json:"data,omitempty"`
			SQL      map[string]string `json:"sql"`
			ExecTime string            `json:"exec_time"`
		}{}

		result.SQL = make(map[string]string)
		for key := range doc.Composition.Subject {
			err = sqlBuilder.AddFilters(custFilters, sqlcomposer.AND)
			if err != nil {
				log.Error(err)
			}

			sqlBuilder.Limit((req.PageIndex-1)*req.PageLimit, req.PageLimit)

			var (
				q string
				a []interface{}
			)
			{
				if sorts.IsEmpty() {
					q, a, err = sqlBuilder.Rebind(key)
				} else {
					q, a, err = sqlBuilder.OrderBy(sorts).Rebind(key)
				}
			}

			if debug == "1" {
				result.SQL[key] = q
			}

			if err != nil {
				log.Error(err)
				c.JSON(http.StatusBadRequest, err)
				return
			}

			if key == "total" {
				var total int64
				err = db.QueryRowx(q, a...).Scan(&total)
				if err != nil {
					log.Error(err)
					c.JSON(http.StatusBadRequest, errJSONWithSQL(err, q))
					return
				}
				result.Total = total
			} else {
				rows, err := db.Queryx(q, a...)

				if err != nil {
					log.Error(err)
					c.JSON(http.StatusBadRequest, errJSONWithSQL(err, q))
					return
				}

				for rows.Next() {
					item := make(map[string]interface{})
					err := rows.MapScan(item)
					if err != nil {
						log.Error(err)
					}

					for k, encoded := range item {
						switch encoded.(type) {
						case []byte:
							item[k] = string(encoded.([]byte))
						}
					}

					result.Data = append(result.Data, item)
				}
			}
		}

		result.ExecTime = time.Since(start).String()

		c.JSON(http.StatusOK, result)
	}
}

type attrsTokenReplacer struct {
	Attrs map[string]string
	DB    *sqlx.DB
}

func (atr *attrsTokenReplacer) TokenReplace(ctx map[string]interface{}) string {
	return ProductAttrsToJoinInStat(atr.DB, atr.Attrs)
}

type attrsFieldsTokenReplacer struct {
	Attrs map[string]string
}

func (atr *attrsFieldsTokenReplacer) TokenReplace(ctx map[string]interface{}) string {
	return ProductAttrsToSelect(atr.Attrs)
}

var once sync.Once

type DictTypesSingleton map[string]string

var dictTypes DictTypesSingleton

type DictType struct {
	SID         string         `db:"sid"`
	Code        string         `db:"code"`
	Name        string         `db:"name"`
	Status      int            `db:"status"`
	Description string         `db:"description"`
	CreateTime  sql.NullString `db:"create_time"`
	UpdateTime  sql.NullString `db:"update_time"`
	IsDelete    int            `db:"is_delete"`
	ParentCode  sql.NullString `db:"parent_code"`
}

func GetMESDictTypes(db *sqlx.DB) DictTypesSingleton {
	once.Do(func() {
		dictTypes = DictTypesSingleton{}

		var types []DictType
		types = []DictType{}
		err := db.Select(&types, "SELECT * FROM fty_dictionary_type")

		if err != nil {
			log.Error(err)
		}

		for _, value := range types {
			dictTypes[value.Code] = value.SID
		}
	})

	return dictTypes
}

func ProductAttrsToJoinInStat(db *sqlx.DB, a map[string]string) string {
	var str []string

	dt := GetMESDictTypes(db)

	for key, _ := range a {
		alias := strings.Replace(key, "-", "_", -1)

		sid, ok := dt[key]

		if !ok {
			os.Exit(500)
		}

		str = append(str,
			fmt.Sprintf(`LEFT JOIN fty_obj_attr AS %s ON %s.attr_sid = '%s' AND %s.obj_sid = fty_product.sid`,
				alias, alias, sid, alias))
	}
	sort.Strings(str)
	return strings.Join(str, " ")
}

func ProductAttrsToSelect(a map[string]string) string {
	var str []string
	for key, value := range a {
		alias := strings.Replace(key, "-", "_", -1)
		str = append(str, fmt.Sprintf("%s.attr_value AS %s", alias, value))
	}
	sort.Strings(str)
	return strings.Join(str, ",")
}

func configureSqlCompose(sb *sqlcomposer.SqlBuilder) {
	sb.RegisterToken("attrs", func(params []sqlcomposer.TokenParam) sqlcomposer.TokenReplacer {
		attrs := map[string]string{}
		for _, p := range params {
			attrs[p.Name] = p.Value
		}
		return &attrsTokenReplacer{
			Attrs: attrs,
			DB:    sb.DB,
		}
	})

	sb.RegisterToken("attrs_fields", func(params []sqlcomposer.TokenParam) sqlcomposer.TokenReplacer {
		attrs := map[string]string{}
		for _, p := range params {
			attrs[p.Name] = p.Value
		}
		return &attrsFieldsTokenReplacer{
			Attrs: attrs,
		}
	})

	_ = sb.RegisterPipelineType("fulltext")
}
