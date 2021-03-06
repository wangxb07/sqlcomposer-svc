package v1

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	log "github.com/sirupsen/logrus"
	"github.com/user/sqlcomposer-svc/models"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"net/http"
	"strconv"
)

var db *sqlx.DB

type Config struct {
	DB *sqlx.DB
}

func Setup(cfg *Config) {
	//init db
	db = cfg.DB
}

func Destroy() {
}

func errJSON(err error) map[string]interface{} {
	return map[string]interface{}{
		"err": err.Error(),
	}
}

func DocListHandler() gin.HandlerFunc {
	return func(context *gin.Context) {
		docs, err := models.Docs().All(context, db)

		if err != nil {
			log.Error(err)
			context.JSON(http.StatusInternalServerError, errJSON(err))
			return
		}

		total, err := models.Docs().Count(context, db)

		if err != nil {
			log.Error(err)
			context.JSON(http.StatusInternalServerError, errJSON(err))
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
		id, err := strconv.Atoi(context.Param("id"))
		if err != nil {
			context.JSON(http.StatusBadRequest, fmt.Sprintf("ID param is required, %s", err))
			return
		}

		docFound, err := models.FindDoc(context, db, id)

		if err != nil {
			log.Error(err)
			context.JSON(http.StatusInternalServerError, errJSON(err))
			return
		}

		context.JSON(http.StatusOK, docFound)
	}
}

func DocUpdateHandler() gin.HandlerFunc {
	return func(context *gin.Context) {
		var (
			id int
		)

		id, err := strconv.Atoi(context.Param("id"))
		if err != nil {
			context.JSON(http.StatusBadRequest, fmt.Sprintf("ID param is required, %s", err))
			return
		}

		docFound, err := models.FindDoc(context, db, id)
		err = context.Bind(&docFound)

		if err != nil {
			log.Error(err)
			context.JSON(http.StatusNotFound, errJSON(err))
			return
		}

		if rowsAff, err := docFound.Update(context, db, boil.Infer()); err != nil {
			log.Error(err)
			context.JSON(http.StatusInternalServerError, errJSON(err))
			return
		} else if rowsAff != 1 {
			log.Error("should only affect one row but affected", rowsAff)
			context.JSON(http.StatusInternalServerError,
				fmt.Sprintf("should only affect one row but affected, %d affected", rowsAff))
			return
		}

		context.JSON(http.StatusOK, docFound)
	}
}

func DocAddHandler() gin.HandlerFunc {
	return func(context *gin.Context) {
		var doc models.Doc
		err := context.Bind(&doc)

		if err != nil {
			log.Error(err)
			context.JSON(http.StatusBadRequest, errJSON(err))
			return
		}

		if err := doc.Insert(boil.WithDebug(context, true), db, boil.Infer()); err != nil {
			log.Error(err)
			context.JSON(http.StatusInternalServerError, errJSON(err))
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
			context.JSON(http.StatusInternalServerError, errJSON(err))
			return
		}

		context.JSON(http.StatusOK, "delete success")
	}
}

func DSNListHandler() gin.HandlerFunc {
	return func(context *gin.Context) {
		res, err := models.DatabaseConfigs().All(context, db)

		if err != nil {
			log.Error(err)
			context.JSON(http.StatusInternalServerError, errJSON(err))
			return
		}

		total, err := models.DatabaseConfigs().Count(context, db)

		if err != nil {
			log.Error(err)
			context.JSON(http.StatusInternalServerError, errJSON(err))
			return
		}

		context.JSON(http.StatusOK, &map[string]interface{}{
			"data":  res,
			"total": total,
		})
	}
}

func DSNGetHandler() gin.HandlerFunc {
	return func(context *gin.Context) {
		id, err := strconv.Atoi(context.Param("id"))
		if err != nil {
			context.JSON(http.StatusBadRequest, fmt.Sprintf("ID param is required, %s", err))
			return
		}

		dbFound, err := models.FindDatabaseConfig(context, db, id)

		if err != nil {
			log.Error(err)
			context.JSON(http.StatusInternalServerError, errJSON(err))
			return
		}

		context.JSON(http.StatusOK, dbFound)
	}
}

func DSNUpdateHandler() gin.HandlerFunc {
	return func(context *gin.Context) {
		var (
			id int
		)

		id, err := strconv.Atoi(context.Param("id"))
		if err != nil {
			context.JSON(http.StatusBadRequest, fmt.Sprintf("ID param is required, %s", err))
			return
		}

		dsnFound, err := models.FindDatabaseConfig(context, db, id)
		err = context.Bind(&dsnFound)

		if err != nil {
			log.Error(err)
			context.JSON(http.StatusNotFound, errJSON(err))
			return
		}

		if rowsAff, err := dsnFound.Update(context, db, boil.Infer()); err != nil {
			log.Error(err)
			context.JSON(http.StatusInternalServerError, errJSON(err))
			return
		} else if rowsAff != 1 {
			log.Error("should only affect one row but affected", rowsAff)
			context.JSON(http.StatusInternalServerError,
				fmt.Sprintf("should only affect one row but affected, %d affected", rowsAff))
			return
		}

		context.JSON(http.StatusOK, dsnFound)
	}
}

func DSNAddHandler() gin.HandlerFunc {
	return func(context *gin.Context) {
		var DSN models.DatabaseConfig
		err := context.Bind(&DSN)

		if err != nil {
			log.Error(err)
			context.JSON(http.StatusBadRequest, errJSON(err))
			return
		}

		if err := DSN.Insert(boil.WithDebug(context, true), db, boil.Infer()); err != nil {
			log.Error(err)
			context.JSON(http.StatusInternalServerError, errJSON(err))
			return
		}

		context.JSON(http.StatusOK, DSN)
	}
}

func DSNDeleteHandler() gin.HandlerFunc {
	return func(context *gin.Context) {
		var id int

		id, err := strconv.Atoi(context.Param("id"))
		if err != nil {
			context.JSON(http.StatusBadRequest, fmt.Sprintf("ID param is required, %s", err))
			return
		}

		dsnFound, err := models.FindDatabaseConfig(context, db, id)
		if err != nil {
			log.Error(err)
			context.JSON(http.StatusNotFound, fmt.Sprintf("not found doc by id %d, %s", id, err))
			return
		}

		if _, err := dsnFound.Delete(context, db); err != nil {
			log.Error(err)
			context.JSON(http.StatusInternalServerError, errJSON(err))
			return
		}

		context.JSON(http.StatusOK, "delete success")
	}
}
