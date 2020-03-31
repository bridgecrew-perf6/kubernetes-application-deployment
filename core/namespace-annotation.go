package core

import (
	"context"
	"errors"
	pb "kubernetes-services-deployment/core/proto"
	"kubernetes-services-deployment/utils"
	"reflect"
)

func (s *Server) AnnotateNamespace(ctx context.Context, request *pb.Namespacerequest) (*pb.Namespaceresponse, error) {
	response := new(pb.Namespaceresponse)
	utils.Info.Println(reflect.TypeOf(ctx))

	if request.CompanyId == "" || request.ProjectId == "" {
		return &pb.Namespaceresponse{}, errors.New("projectId or companyId must not be empty")
	}

	agent, err := GetGrpcAgentConnection()
	if err != nil {
		utils.Error.Println(err)
		return &pb.Namespaceresponse{}, err
	}

	err = agent.InitializeAgentClient(request.ProjectId, request.CompanyId)
	if err != nil {
		utils.Error.Println(err)
		return &pb.Namespaceresponse{}, err
	}

	defer agent.connection.Close()

	resp, err := agent.AddLabel(ctx, request)
	if err != nil {
		return &pb.Namespaceresponse{}, err
	}

	response.Message = resp
	return response, nil

}
