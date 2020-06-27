package restapi

import (
	"github.com/gin-gonic/gin"
	"github.com/user/sqlcomposer-svc/restapi/v1"
	"net/http"
	"time"
)

func InitRoutes() *gin.Engine {
	router := gin.Default()
	router.GET("/", func(c *gin.Context) {
		time.Sleep(5 * time.Second)
		c.String(http.StatusOK, "Welcome Gin Server")
	})

	// Ping test
	router.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	// group v1 for mgt
	rv1 := router.Group("/v1")
	{
		rv1.GET("/doc", v1.DocListHandler())
		rv1.GET("/doc/:id", v1.DocGetHandler())
		rv1.PATCH("/doc/:id", v1.DocUpdateHandler())
		rv1.POST("/doc", v1.DocAddHandler())
		rv1.DELETE("/doc/:id", v1.DocDeleteHandler())

		rv1.GET("/dsn", v1.DSNListHandler())
		rv1.GET("/dsn/:id", v1.DSNGetHandler())
		rv1.PATCH("/dsn/:id", v1.DSNUpdateHandler())
		rv1.POST("/dsn", v1.DSNAddHandler())
		rv1.DELETE("/dsn/:id", v1.DSNDeleteHandler())
	}

	return router
}
