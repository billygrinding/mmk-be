package main

import (
	service "github.com/billygrinding/mmk-be/app/healthcheck"
	"github.com/billygrinding/mmk-be/pb"
	"google.golang.org/grpc"
	"log"
	"net"
)

const (
	// Port for gRPC server to listen to
	PORT = ":2000"
)

func main() {

	lis, err := net.Listen("tcp", PORT)

	if err != nil {
		log.Fatalf("failed connection: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterHealthCheckServiceServer(s, &service.HealthcheckServer{})

	log.Printf("Server listening at %v", lis.Addr())

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to server: %v", err)
	}

}
