package v1

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	log "github.com/sirupsen/logrus"
	"github.com/user/sqlcomposer-svc/models"
	"github.com/volatiletech/sqlboiler/boil"
	"net/http"
	"strconv"
)

var db *sqlx.DB

type Config struct {
	DNS      string
	LogLevel log.Level
}

func Setup(cfg *Config) {
	//init db
	db = sqlx.MustConnect("mysql", cfg.DNS)
}

func DocListHandler() gin.HandlerFunc {
	return func(context *gin.Context) {
		docs, err := models.Docs().All(context, db)

		if err != nil {
			log.Error(err)
			context.JSON(http.StatusInternalServerError, err)
			return
		}

		total, err := models.Docs().Count(context, db)

		if err != nil {
			log.Error(err)
			context.JSON(http.StatusInternalServerError, err)
			return
		}

		context.JSON(http.StatusOK, &map[string]interface{}{
			"data":  docs,
			"total": total,
		})
	}
}

func DocGetHandler() gin.HandlerFunc {
	return func(context *gin.Context) {
		doc, err := models.Docs().One(context, db)

		if err != nil {
			log.Error(err)
			context.JSON(http.StatusInternalServerError, err)
			return
		}

		context.JSON(http.StatusOK, doc)
	}
}

func DocUpdateHandler() gin.HandlerFunc {
	return func(context *gin.Context) {
		var doc models.Doc
		err := context.Bind(&doc)

		if err != nil {
			log.Error(err)
			context.JSON(http.StatusBadRequest, err)
			return
		}

		if rowsAff, err := doc.Update(context, db, boil.Infer()); err != nil {
			log.Error(err)
			context.JSON(http.StatusInternalServerError, err)
			return
		} else if rowsAff != 1 {
			log.Error("should only affect one row but affected", rowsAff)
			context.JSON(http.StatusInternalServerError, err)
			return
		}

		context.JSON(http.StatusOK, doc)
	}
}

func DocAddHandler() gin.HandlerFunc {
	return func(context *gin.Context) {
		var doc models.Doc
		err := context.Bind(&doc)

		if err != nil {
			log.Error(err)
			context.JSON(http.StatusBadRequest, err)
			return
		}

		if err := doc.Insert(context, db, boil.Infer()); err != nil {
			log.Error(err)
			context.JSON(http.StatusInternalServerError, err)
			return
		}

		context.JSON(http.StatusOK, doc)
	}
}

func DocDeleteHandler() gin.HandlerFunc {
	return func(context *gin.Context) {
		var id int

		id, err := strconv.Atoi(context.Param("id"))
		if err != nil {
			context.JSON(http.StatusBadRequest, fmt.Sprintf("ID param is required, %s", err))
			return
		}

		docFound, err := models.FindDoc(context, db, id)
		if err != nil {
			log.Error(err)
			context.JSON(http.StatusNotFound, fmt.Sprintf("not found doc by id %d, %s", id, err))
			return
		}

		if _, err := docFound.Delete(context, db); err != nil {
			log.Error(err)
			context.JSON(http.StatusInternalServerError, err)
			return
		}

		context.JSON(http.StatusOK, "delete success")
	}
}