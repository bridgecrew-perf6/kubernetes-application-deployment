package core

import (
	pb "bitbucket.org/cloudplex-devs/kubernetes-services-deployment/core/proto"
	"bitbucket.org/cloudplex-devs/kubernetes-services-deployment/utils"
	"context"
	"errors"
	"reflect"
)

func (s *Server) GetK8SResource(ctx context.Context, request *pb.KubernetesResourceRequest) (response *pb.KubernetesResourceResponse, err error) {
	response = new(pb.KubernetesResourceResponse)
	utils.Info.Println(reflect.TypeOf(ctx))

	if request.CompanyId == "" || request.InfraId == "" {
		return &pb.KubernetesResourceResponse{}, errors.New("projectId or companyId must not be empty")
	}

	agent, err := GetGrpcAgentConnection()
	if err != nil {
		utils.Error.Println(err)
		return &pb.KubernetesResourceResponse{}, err
	}
	agent.InfraId = request.InfraId
	agent.CompanyId = request.CompanyId

	err = agent.InitializeAgentClient()
	if err != nil {
		utils.Error.Println(err)
		return &pb.KubernetesResourceResponse{}, err
	}

	defer agent.Connection.Close()

	resp, err := agent.GetK8sResources(ctx, request)
	if err != nil {
		return &pb.KubernetesResourceResponse{}, err
	}

	response.Resource = resp
	return response, err

}

//func convertToK8sStruct(resourceData []byte) error {
//	decode := scheme.Codecs.UniversalDeserializer().Decode
//	obj, _, err := decode(resourceData, nil, nil)
//	if err != nil {
//		utils.Error.Println(fmt.Sprintf("Error while decoding YAML object. Err was: %s", err))
//		return errors.New(fmt.Sprintf("Error while decoding YAML object. Err was: %s", err))
//	}
//	switch o := obj.(type) {
//	case *v2.Deployment:
//		var deployment v2.Deployment
//		err := json.Unmarshal(resourceData, &deployment)
//	case *v2.DaemonSet:
//	case *v2.ReplicaSet:
//
//	}
//
//	return nil
//}
