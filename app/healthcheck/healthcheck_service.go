package service

import (
	"context"
	"github.com/billygrinding/mmk-be/pb"
	"log"
)

type HealthcheckServer struct {
	pb.HealthCheckServiceServer
}

func (s *HealthcheckServer) HealthCheck(ctx context.Context, in *pb.Request) (*pb.Response, error) {
	log.Printf("Received: %v", in.GetValue())
	response := &pb.Response{
		Value: pb.Status_OK,
	}

	return response, nil

}
