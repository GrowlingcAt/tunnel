package main

import (
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"user/controller/login"
	data2 "user/data"
	"user/middleware"
	"user/pkg/config"
	"user/pkg/db/mysql"
	"user/pkg/db/redis"
	"user/pkg/log"
	"user/pkg/storage/cos"
	"user/routers"
)

var configFile = flag.String("config", "dev.config.yaml", "")

func main() {
	flag.Parse()
	//初始化配置文件
	config.InitConfig(*configFile)
	cnf := config.GetConfig()

	log.SetLevel(cnf.Log.Level)
	log.SetOutput(log.GetRotateWriter(cnf.Log.LogPath))
	log.SetPrintCaller(true)

	logger := log.NewLogger()
	logger.SetOutput(log.GetRotateWriter(cnf.Log.LogPath))
	logger.SetLevel(cnf.Log.Level)
	logger.SetPrintCaller(true)

	gin.SetMode(cnf.Http.Mode)
	r := gin.Default()
	r.Use(middleware.Cors())
	r.GET("/health", func(*gin.Context) {})
	api := r.Group("/api")

	mysql.InitMysql(cnf)
	redis.InitRedisPool(cnf)
	sf := cos.NewCosStorageFactory(cnf.Cos.BucketUrl, cnf.Cos.SecretId, cnf.Cos.SecretKey, cnf.Cos.CDNDomain)
	redisPool := redis.GetPool()
	data := data2.NewData(mysql.GetDB())
	loginController := login.NewLoginController(logger, cnf, redisPool, sf, data)

	routers.InitLoginRouters(api, loginController)

	fs := http.FileServer(http.Dir("www"))
	r.NoRoute(func(ctx *gin.Context) {
		fs.ServeHTTP(ctx.Writer, ctx.Request)
	})
	r.GET("/", func(ctx *gin.Context) {
		http.ServeFile(ctx.Writer, ctx.Request, "www/index.html")
	})

	err := r.Run(fmt.Sprintf("%s:%d", cnf.Http.IP, cnf.Http.Port))
	if err != nil {
		log.Error(err)
	}
}
