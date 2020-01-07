package core

import (
	"context"
	"kubernetes-services-deployment/constants"
	pb "kubernetes-services-deployment/core/proto"
	"kubernetes-services-deployment/utils"
	"reflect"
)

type Server struct {
}

func (s *Server) CreateService(ctx context.Context, request *pb.ServiceRequest) (response *pb.SerivceFResponse, err error) {
	response = new(pb.SerivceFResponse)
	utils.Info.Println(reflect.TypeOf(ctx))
	cpCtx := &Context{}
	cpCtx.Keys = make(map[string]interface{})
	cpCtx.Keys["token"] = request.Token
	cpCtx.Keys["companyId"] = request.CompanyId

	conn, err := GetGrpcAgentConnection()
	if err != nil {
		utils.Error.Println(err)
		return response, err
	}

	service, err := conn.AgentCrdManager(constants.POST, request)
	if err != nil {
		utils.Error.Println(err)
		return response, err
	}

	utils.Info.Println(string(service))
	response.Service = service
	return response, nil
}
func (s *Server) GetService(ctx context.Context, request *pb.ServiceRequest) (response *pb.SerivceFResponse, err error) {
	response = new(pb.SerivceFResponse)
	utils.Info.Println(reflect.TypeOf(ctx))
	cpCtx := &Context{}
	cpCtx.Keys = make(map[string]interface{})
	cpCtx.Keys["token"] = request.Token
	cpCtx.Keys["companyId"] = request.CompanyId

	conn, err := GetGrpcAgentConnection()
	if err != nil {
		utils.Error.Println(err)
		return response, err
	}

	service, err := conn.AgentCrdManager(constants.GET, request)
	if err != nil {
		utils.Error.Println(err)
		return response, err
	}

	utils.Info.Println(string(service))
	response.Service = service
	return response, nil
}
func (s *Server) DeleteService(ctx context.Context, request *pb.ServiceRequest) (response *pb.SerivceFResponse, err error) {
	response = new(pb.SerivceFResponse)
	utils.Info.Println(reflect.TypeOf(ctx))
	cpCtx := &Context{}
	cpCtx.Keys = make(map[string]interface{})
	cpCtx.Keys["token"] = request.Token
	cpCtx.Keys["companyId"] = request.CompanyId

	conn, err := GetGrpcAgentConnection()
	if err != nil {
		utils.Error.Println(err)
		return response, err
	}

	service, err := conn.AgentCrdManager(constants.DELETE, request)
	if err != nil {
		utils.Error.Println(err)
		return response, err
	}

	utils.Info.Println(string(service))
	response.Service = service
	return response, nil
}
func (s *Server) PatchService(ctx context.Context, request *pb.ServiceRequest) (response *pb.SerivceFResponse, err error) {
	response = new(pb.SerivceFResponse)
	utils.Info.Println(reflect.TypeOf(ctx))
	cpCtx := &Context{}
	cpCtx.Keys = make(map[string]interface{})
	cpCtx.Keys["token"] = request.Token
	cpCtx.Keys["companyId"] = request.CompanyId

	conn, err := GetGrpcAgentConnection()
	if err != nil {
		utils.Error.Println(err)
		return response, err
	}

	service, err := conn.AgentCrdManager(constants.PATCH, request)
	if err != nil {
		utils.Error.Println(err)
		return response, err
	}

	utils.Info.Println(string(service))
	response.Service = service
	return response, nil
}
func (s *Server) PutService(ctx context.Context, request *pb.ServiceRequest) (response *pb.SerivceFResponse, err error) {
	response = new(pb.SerivceFResponse)
	utils.Info.Println(reflect.TypeOf(ctx))
	cpCtx := &Context{}
	cpCtx.Keys = make(map[string]interface{})
	cpCtx.Keys["token"] = request.Token
	cpCtx.Keys["companyId"] = request.CompanyId

	conn, err := GetGrpcAgentConnection()
	if err != nil {
		utils.Error.Println(err)
		return response, err
	}

	service, err := conn.AgentCrdManager(constants.PUT, request)
	if err != nil {
		utils.Error.Println(err)
		return response, err
	}

	utils.Info.Println(string(service))
	response.Service = service

	//c, err := GetKubernetesClient(cpCtx, &request.ProjectId)
	//if err != nil {
	//	utils.Error.Println(err)
	//	return response, err
	//}
	//var req interface{}
	//err = json.Unmarshal(request.Service, &req)
	//if err != nil {
	//	utils.Error.Println(err)
	//	return response, err
	//}
	//responseObj, _ := c.crdManager(req, constants.PUT)
	//if responseObj.Error != "" {
	//	utils.Error.Println(err)
	//	return response, err
	//}
	//raw, err := json.Marshal(responseObj.Data)
	//if responseObj.Error != "" {
	//	utils.Error.Println(err)
	//	return response, err
	//}
	//response.Service = raw
	return response, nil
}
