package main

import (
	"crypto/tls"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func main() {
	// 加载证书和私钥
	certificate, err := tls.LoadX509KeyPair("server.crt", "server.key")
	if err != nil {
		log.Fatalf("Failed to load certificate and key: %v", err)
	}

	// 创建一个新的 TLS 配置
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{certificate},
	}

	// 创建一个新的 gRPC 服务器
	grpcServer := grpc.NewServer(grpc.Creds(credentials.NewTLS(tlsConfig)))

	// 注册 gRPC 服务
	// ...

	// 监听网络地址
	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	// 启动 gRPC 服务器
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Failed to start gRPC server: %v", err)
	}
}
