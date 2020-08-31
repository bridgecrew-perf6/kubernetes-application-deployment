package core

import (
	pb "bitbucket.org/cloudplex-devs/kubernetes-services-deployment/core/proto"
	"bitbucket.org/cloudplex-devs/kubernetes-services-deployment/utils"
	"context"
	"errors"
	"reflect"
)

func (s *Server) AnnotateNamespace(ctx context.Context, request *pb.Namespacerequest) (*pb.Namespaceresponse, error) {
	response := new(pb.Namespaceresponse)
	utils.Info.Println(reflect.TypeOf(ctx))

	if request.CompanyId == "" || request.InfraId == "" {
		return &pb.Namespaceresponse{}, errors.New("projectId or companyId must not be empty")
	}

	agent, err := GetGrpcAgentConnection()
	if err != nil {
		utils.Error.Println(err)
		return &pb.Namespaceresponse{}, err
	}

	err = agent.InitializeAgentClient()
	if err != nil {
		utils.Error.Println(err)
		return &pb.Namespaceresponse{}, err
	}

	defer agent.Connection.Close()

	resp, err := agent.AddLabel(ctx, request)
	if err != nil {
		return &pb.Namespaceresponse{}, err
	}

	response.Message = resp
	return response, nil

}
