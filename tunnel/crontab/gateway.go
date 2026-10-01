package crontab

import (
	"tunnel/data"
	nginx_gateway "tunnel/nginx-gateway"
	"tunnel/pkg/config"
	"tunnel/pkg/log"
)

type gatewayJob struct {
	config  *config.Config
	log     log.ILogger
	gateway *nginx_gateway.Gateway
}

func NewGatewayJob(cnf *config.Config, log log.ILogger, data data.IData) *gatewayJob {
	return &gatewayJob{
		config:  cnf,
		log:     log,
		gateway: nginx_gateway.NewGateway(cnf, log, data),
	}
}
func (j *gatewayJob) Run() {
	err := j.gateway.DeployK8s()
	if err != nil {
		j.log.Error(err)
		return
	}
}
