package crontab

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"sync"
	"user/pkg/config"
	"user/pkg/grpc_client_pool"
	"user/pkg/log"
	"user/pkg/zerror"
)

var pool grpc_client_pool.ClientPool
var once sync.Once

func NewCrontabClientPool() grpc_client_pool.ClientPool {
	var err error
	if pool != nil {
		return pool
	}
	once.Do(func() {
		cnf := config.GetConfig()
		pool, err = grpc_client_pool.NewPool(cnf.DependOn.Crontab.Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Error(zerror.NewByErr(err))
		}
	})
	return pool
}
