package server

import "google.golang.org/grpc"

func GetOptions() (opts []grpc.ServerOption) {
	opts = make([]grpc.ServerOption, 0)
	opts = append(opts, GetKeepaliveOpt()...)
	opts = append(opts, grpc.StreamInterceptor(StreamInterceptor))
	opts = append(opts, grpc.UnaryInterceptor(UnaryInterceptor))
	return opts
}
