package routers

import (
	"github.com/gin-gonic/gin"
	"tunnel/controller"
)

func InitAppRouters(api *gin.RouterGroup, c *controller.App) {
	v1 := api.Group("/v1")
	v1.POST("/app", c.Add)
	v1.PUT("/app", c.Update)
	v1.GET("/app", c.List)
	v1.POST("/app/deploy", c.DeployServer)
	v1.DELETE("/app", c.Delete)
}
