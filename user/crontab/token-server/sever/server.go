package server

import (
	"context"
	"crontab/pkg/config"
	"crontab/pkg/log"
	"crontab/pkg/wx-api"
	"crontab/pkg/wx-api/wx-official"
	"crontab/proto"
)

type tokenServer struct {
	proto.UnimplementedTokenServer
	config *config.Config
	log    log.ILogger
}

func NewTokenServer(config *config.Config, log log.ILogger) proto.TokenServer {
	return &tokenServer{
		config: config,
		log:    log,
	}
}

func (s *tokenServer) GetToken(ctx context.Context, in *proto.TokenRequest) (*proto.TokenResponse, error) {
	var token wx_api.Token
	secret := s.getSecret(in)
	token = wx_official.NewWxOfficial(in.Id, secret)
	if token != nil {
		accessToken, err := token.GetToken()
		if err != nil {
			s.log.Error(err)
			return nil, err
		}
		res := &proto.TokenResponse{
			AccessToken: accessToken.AccessToken,
		}
		return res, err
	}
	return nil, nil
}

func (s *tokenServer) getSecret(in *proto.TokenRequest) string {
	for _, item := range s.config.WxOfficials {
		if item.AppId == in.Id {
			return item.Secret
		}
	}
	return ""
}
