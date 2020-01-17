package core

import (
	"context"
	"errors"
	pb "kubernetes-services-deployment/core/proto"
	"kubernetes-services-deployment/utils"
	"reflect"
)

func (s *Server) GetK8SResource(ctx context.Context, request *pb.K8SResourceRequest) (response *pb.K8SResourceResponse, err error) {
	response = new(pb.K8SResourceResponse)
	utils.Info.Println(reflect.TypeOf(ctx))

	if request.CompanyId == "" || request.ProjectId == "" {
		return &pb.K8SResourceResponse{}, errors.New("projectId or companyId must not be empty")
	}

	conn, err := GetGrpcAgentConnection()
	if err != nil {
		utils.Error.Println(err)
		return &pb.K8SResourceResponse{}, err
	}

	resp, err := conn.GetK8sResources(ctx, request)
	if err != nil {
		return &pb.K8SResourceResponse{}, err
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
