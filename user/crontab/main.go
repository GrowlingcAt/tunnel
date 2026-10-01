package main

import (
	"context"
	"crontab/cron"
	"crontab/pkg/config"
	"crontab/pkg/db/redis"
	"crontab/pkg/log"
	token_server "crontab/token-server"
	"flag"
	"os"
	"os/signal"
)

var (
	configFile = flag.String("config", "dev.config.yaml", "")
)

func main() {

	flag.Parse()
	config.InitConf(*configFile)
	cnf := config.GetConf()

	logger := log.NewLogger()
	logger.SetOutput(log.GetRotateWriter(cnf.Log.LogPath))
	logger.SetLevel(cnf.Log.Level)
	logger.SetPrintCaller(true)

	log.SetLevel(cnf.Log.Level)
	log.SetOutput(log.GetRotateWriter(cnf.Log.LogPath))
	log.SetPrintCaller(true)

	//初始化redis
	redis.InitRedisPool()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer stop()

	go cron.Run()

	go token_server.Start(cnf, logger)

	<-ctx.Done()

}
