package core

import (
	"context"
	"errors"
	pb "kubernetes-services-deployment/core/proto"
	"kubernetes-services-deployment/utils"
	"reflect"
)

func (s *Server) KillPod(ctx context.Context, request *pb.PodRequest) (*pb.PodResponse, error) {
	response := new(pb.PodResponse)
	utils.Info.Println(reflect.TypeOf(ctx))

	if request.CompanyId == "" || request.ProjectId == "" {
		return &pb.PodResponse{}, errors.New("projectId or companyId must not be empty")
	}

	agent, err := GetGrpcAgentConnection()
	if err != nil {
		utils.Error.Println(err)
		return &pb.PodResponse{}, err
	}

	err = agent.InitializeAgentClient(request.ProjectId, request.CompanyId)
	if err != nil {
		utils.Error.Println(err)
		return &pb.PodResponse{}, err
	}

	defer agent.connection.Close()

	resp, err := agent.Killingpod(ctx, request)
	if err != nil {
		return &pb.PodResponse{}, err
	}
	response.Message = resp
	return response, nil
}
