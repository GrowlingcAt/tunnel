package token_server

import (
	"crontab/pkg/config"
	"crontab/pkg/log"
	"crontab/proto"
	server "crontab/token-server/sever"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	grpchealth "google.golang.org/grpc/health/grpc_health_v1"
	"net"
)

func Start(cnf *config.Config, logger log.ILogger) {
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", cnf.Server.Host, cnf.Server.Port))
	if err != nil {
		log.Error(err)
		panic(err)
	}
	s := grpc.NewServer(server.GetOptions()...)

	tokenServer := server.NewTokenServer(cnf, logger)
	proto.RegisterTokenServer(s, tokenServer)

	//添加健康检查
	healthCheck := health.NewServer()
	grpchealth.RegisterHealthServer(s, healthCheck)

	if err := s.Serve(lis); err != nil {
		panic(err)
	}

}
