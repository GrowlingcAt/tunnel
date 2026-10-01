package server

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"time"
)

func GetKeepaliveOpt() (opt []grpc.ServerOption) {
	//服务端强制保活策略,客户端违反该策略则强制关闭连接
	var kaep = keepalive.EnforcementPolicy{
		//客户端ping服务器，最小时间间隔，小于该时间间隔则强制关闭连接
		MinTime: 10 * time.Second,
		//当没有任何活动流的情况下，是否允许被ping
		PermitWithoutStream: true,
	}

	var kasp = keepalive.ServerParameters{
		//客户端空闲10分钟发送goaway指令（尝试断开连接）
		MaxConnectionIdle: 600 * time.Second,
		//最大连接时长30分钟，超时发送goaway
		MaxConnectionAge: 1800 * time.Second,
		//强制关闭前等待时长
		MaxConnectionAgeGrace: 5 * time.Second,
		//客户端空闲50s，发送ping保活
		Time: 50 * time.Second,
		//ping ack 1s内没有返回则认定连接断开
		Timeout: 1 * time.Second,
	}
	return []grpc.ServerOption{grpc.KeepaliveEnforcementPolicy(kaep), grpc.KeepaliveParams(kasp)}
}
