package main

import (
	"context"
	"github.com/billygrinding/mmk-be/pb"
	"google.golang.org/grpc/grpclog"
	"io/ioutil"
	"log"
	"net"
	"os"

	"google.golang.org/grpc"
)

const (
	// Port for gRPC server to listen to
	PORT = ":2000"
)

type HealthcheckServer struct {
	pb.HealthCheckServiceServer
}

func (s *HealthcheckServer) CreateHealthCheck(ctx context.Context, in *pb.Request) (*pb.Response, error) {
	log.Printf("Received: %v", in.GetValue())
	response := &pb.Response{
		Value: pb.Status_OK,
	}

	return response, nil

}

func main() {

	lis, err := net.Listen("tcp", PORT)

	if err != nil {
		log.Fatalf("failed connection: %v", err)
	}

	s := grpc.NewServer()

	grpclog.SetLoggerV2(grpclog.NewLoggerV2(os.Stdout, ioutil.Discard, ioutil.Discard))
	pb.RegisterHealthCheckServiceServer(s, &HealthcheckServer{})

	log.Printf("Server listening at %v", lis.Addr())

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to server: %v", err)
	}

}
