package test_case

import (
	"context"
	"crontab/pkg/config"
	"crontab/pkg/db/redis"
	"crontab/pkg/log"
	"crontab/proto"
	server "crontab/token-server/sever"
	"flag"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"net"
	"testing"
)

var configPath = flag.String("config", "../config.yaml", "单元测试配置文件")

func TestMain(m *testing.M) {
	flag.Parse()
	config.InitConf(*configPath)
	redis.InitRedisPool()

	//启动grpc server
	go startGRPCServer()
	m.Run()
}

func startGRPCServer() {
	cnf := config.GetConf()
	lis, err := net.Listen("tcp", "localhost:50051")
	if err != nil {
		panic(err)
	}
	s := grpc.NewServer(getServerOptions()...)
	logger := log.NewLogger()
	logger.SetLevel(cnf.Log.Level)
	logger.SetPrintCaller(true)
	tokenServer := server.NewTokenServer(cnf, logger)
	proto.RegisterTokenServer(s, tokenServer)

	if err := s.Serve(lis); err != nil {
		panic(err)
	}
}

func getServerOptions() []grpc.ServerOption {
	var opts = make([]grpc.ServerOption, 0)
	opts = append(opts, server.GetKeepaliveOpt()...)
	opts = append(opts, grpc.StreamInterceptor(server.StreamInterceptor))
	opts = append(opts, grpc.UnaryInterceptor(server.UnaryInterceptor))
	return opts
}

func getOptions() []grpc.DialOption {
	opts := make([]grpc.DialOption, 0)
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	return opts
}
func appendBearerTokenToContext(ctx context.Context) context.Context {
	token := config.GetConf().Server.AccessToken
	md := metadata.Pairs("authorization", "Bearer "+token)
	return metadata.NewOutgoingContext(ctx, md)

}

func TestGetWxOfficialToken(t *testing.T) {
	conn, err := grpc.Dial("localhost:50051", getOptions()...)
	if err != nil {
		t.Error(err)
		return
	}
	defer conn.Close()
	client := proto.NewTokenClient(conn)
	in := &proto.TokenRequest{
		Id:  config.GetConf().WxOfficials[0].AppId,
		App: "",
		Typ: proto.TokenType_WECHATOFFICIAL,
	}
	ctx := context.Background()
	ctx = appendBearerTokenToContext(ctx)
	res, err := client.GetToken(ctx, in)
	if err != nil {
		t.Error(err)
		return
	}
	if res.AccessToken == "" {
		t.Error("access_token获取失败")
		return
	}
	t.Log(res.AccessToken)
}

func TestGetWeComToken(t *testing.T) {
	conn, err := grpc.Dial("localhost:50051", getOptions()...)
	if err != nil {
		t.Error(err)
		return
	}
	defer conn.Close()
	client := proto.NewTokenClient(conn)
	in := &proto.TokenRequest{
		Id:  config.GetConf().WeComs[0].CorpId,
		App: config.GetConf().WeComs[0].App,
		Typ: proto.TokenType_WECOM,
	}
	ctx := context.Background()
	ctx = appendBearerTokenToContext(ctx)
	res, err := client.GetToken(ctx, in)
	if err != nil {
		t.Error(err)
		return
	}
	if res.AccessToken == "" {
		t.Error("access_token获取失败")
		return
	}
	t.Log(res.AccessToken)
}
