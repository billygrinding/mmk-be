package main

import (
	"github.com/billygrinding/mmk-be/app/config"
	service "github.com/billygrinding/mmk-be/app/healthcheck"
	"github.com/billygrinding/mmk-be/pb"
	"google.golang.org/grpc"
	"log"
	"net"
)

func main() {
	conf := initConfigReader()

	lis, err := net.Listen("tcp", conf.App.Port)
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

func initConfigReader() config.Root {
	rootConfig := config.Load(".env")
	return rootConfig
}
