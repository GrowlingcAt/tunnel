package server

import (
	"context"
	"crontab/pkg/config"
	"errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"strings"
)

func StreamInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	err := auth(ss.Context())
	if err != nil {
		return err
	}
	return handler(srv, ss)
}

func UnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	if info.FullMethod != "/grpc.health.v1.Health/Check" {
		err = auth(ctx)
		if err != nil {
			return nil, err
		}
	}
	return handler(ctx, req)
}

func auth(ctx context.Context) error {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return errors.New("获取元数据失败，身份认证失败")
	}
	authorization := md["authorization"]
	if !valid(authorization) {
		return errors.New("身份令牌校验失败，身份认证失败")
	}
	return nil
}

func valid(authorization []string) bool {
	if len(authorization) < 1 {
		return false
	}
	token := strings.TrimPrefix(authorization[0], "Bearer ")
	return token == config.GetConf().Server.AccessToken
}
