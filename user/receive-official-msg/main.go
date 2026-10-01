package main

import (
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	"receive-official-msg/controller"
	"receive-official-msg/pkg/config"
)

var configFile = flag.String("config", "dev.config.yaml", "")

func main() {
	flag.Parse()
	config.InitConf(*configFile)
	cnf := config.GetConf()
	r := gin.Default()
	r.GET("/api/v1/login/official/receive", controller.WxCheckSignature)
	r.POST("/api/v1/login/official/receive", controller.WxReceiveMessage)
	r.Run(fmt.Sprintf("%s:%d", cnf.Http.IP, cnf.Http.Port))
}
