package main

import (
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"tunnel/controller"
	data2 "tunnel/data"
	"tunnel/middleware"
	"tunnel/pkg/config"
	"tunnel/pkg/db/mysql"
	"tunnel/pkg/db/redis"
	domain_resolution "tunnel/pkg/domain-resolution"
	"tunnel/pkg/k8s"
	"tunnel/pkg/log"
	"tunnel/routers"
)

var (
	configFile = flag.String("config", "dev.config.yaml", "Path to the configuration file")
	// k8s-conf
	k8sConf = flag.String("k8s-conf", "", "Path to the Kubernetes configuration file")
)

func main() {
	flag.Parse()
	config.InitConfig(*configFile)
	cnf := config.GetConfig()
	mysql.InitMysql()
	redis.InitRedis()

	log.SetLevel(cnf.Log.Level)
	log.SetOutput(log.GetRotateWriter(cnf.Log.LogPath))
	log.SetPrintCaller(true)

	logger := log.NewLogger()
	logger.SetLevel(cnf.Log.Level)
	logger.SetOutput(log.GetRotateWriter(cnf.Log.LogPath))
	logger.SetPrintCaller(true)

	k8s.InitK8sClient(*k8sConf)
	data := data2.NewData(mysql.GetDB())

	dnsFactory := domain_resolution.NewDnsFactory(cnf)

	appController := controller.NewApp(cnf, logger, data, dnsFactory)
	gin.SetMode(cnf.Http.Mode)
	r := gin.Default()
	r.Use(middleware.Cors())
	fs := http.FileServer(http.Dir("www"))
	r.NoRoute(func(ctx *gin.Context) {
		fs.ServeHTTP(ctx.Writer, ctx.Request)
	})
	r.GET("/", func(ctx *gin.Context) {
		http.ServeFile(ctx.Writer, ctx.Request, "www/index.html")
	})
	r.GET("/health", func(ctx *gin.Context) {})

	api := r.Group("/api")
	api.Use(middleware.Auth())
	routers.InitAppRouters(api, appController)
	r.Run(fmt.Sprintf("%s:%d", cnf.Http.IP, cnf.Http.Port))
}
