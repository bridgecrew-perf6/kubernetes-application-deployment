package core

import (
	"context"
	"encoding/json"
	"errors"
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
	c, err := GetKubernetesClient(cpCtx, &request.ProjectId)
	if err != nil {
		utils.Error.Println(err)
		return response, err
	}
	var req interface{}
	err = json.Unmarshal(request.Service, &req)
	if err != nil {
		utils.Error.Println(err)
		return response, err
	}
	responseObj, _ := c.crdManager(req, constants.POST)
	if responseObj.Error != "" {
		utils.Error.Println(responseObj.Error)
		return response, errors.New(responseObj.Error)
	}
	raw, err := json.Marshal(responseObj.Data)
	if responseObj.Error != "" {
		utils.Error.Println(err)
		return response, err
	}
	response.Service = raw
	return response, nil

}
func (s *Server) GetService(ctx context.Context, request *pb.ServiceRequest) (response *pb.SerivceFResponse, err error) {
	response = new(pb.SerivceFResponse)
	utils.Info.Println(reflect.TypeOf(ctx))
	cpCtx := &Context{}
	cpCtx.Keys = make(map[string]interface{})
	cpCtx.Keys["token"] = request.Token
	cpCtx.Keys["companyId"] = request.CompanyId
	c, err := GetKubernetesClient(cpCtx, &request.ProjectId)
	if err != nil {
		utils.Error.Println(err)
		return response, err
	}
	var req interface{}
	err = json.Unmarshal(request.Service, &req)
	if err != nil {
		utils.Error.Println(err)
		return response, err
	}
	responseObj, _ := c.crdManager(req, constants.GET)
	if responseObj.Error != "" {
		utils.Error.Println(err)
		return response, err
	}
	raw, err := json.Marshal(responseObj.Data)
	if responseObj.Error != "" {
		utils.Error.Println(err)
		return response, err
	}
	response.Service = raw
	return response, nil
}
func (s *Server) DeleteService(ctx context.Context, request *pb.ServiceRequest) (response *pb.SerivceFResponse, err error) {
	response = new(pb.SerivceFResponse)
	utils.Info.Println(reflect.TypeOf(ctx))
	cpCtx := &Context{}
	cpCtx.Keys = make(map[string]interface{})
	cpCtx.Keys["token"] = request.Token
	cpCtx.Keys["companyId"] = request.CompanyId
	c, err := GetKubernetesClient(cpCtx, &request.ProjectId)
	if err != nil {
		utils.Error.Println(err)
		return response, err
	}
	var req interface{}
	err = json.Unmarshal(request.Service, &req)
	if err != nil {
		utils.Error.Println(err)
		return response, err
	}
	responseObj, _ := c.crdManager(req, constants.DELETE)
	if responseObj.Error != "" {
		utils.Error.Println(err)
		return response, err
	}
	raw, err := json.Marshal(responseObj.Data)
	if responseObj.Error != "" {
		utils.Error.Println(err)
		return response, err
	}
	response.Service = raw
	return response, nil
}
func (s *Server) PatchService(ctx context.Context, request *pb.ServiceRequest) (response *pb.SerivceFResponse, err error) {
	response = new(pb.SerivceFResponse)
	utils.Info.Println(reflect.TypeOf(ctx))
	cpCtx := &Context{}
	cpCtx.Keys = make(map[string]interface{})
	cpCtx.Keys["token"] = request.Token
	cpCtx.Keys["companyId"] = request.CompanyId
	c, err := GetKubernetesClient(cpCtx, &request.ProjectId)
	if err != nil {
		utils.Error.Println(err)
		return response, err
	}
	var req interface{}
	err = json.Unmarshal(request.Service, &req)
	if err != nil {
		utils.Error.Println(err)
		return response, err
	}
	responseObj, _ := c.crdManager(req, constants.PATCH)
	if responseObj.Error != "" {
		utils.Error.Println(err)
		return response, err
	}
	raw, err := json.Marshal(responseObj.Data)
	if responseObj.Error != "" {
		utils.Error.Println(err)
		return response, err
	}
	response.Service = raw
	return response, nil
}
func (s *Server) PutService(ctx context.Context, request *pb.ServiceRequest) (response *pb.SerivceFResponse, err error) {
	response = new(pb.SerivceFResponse)
	utils.Info.Println(reflect.TypeOf(ctx))
	cpCtx := &Context{}
	cpCtx.Keys = make(map[string]interface{})
	cpCtx.Keys["token"] = request.Token
	cpCtx.Keys["companyId"] = request.CompanyId
	c, err := GetKubernetesClient(cpCtx, &request.ProjectId)
	if err != nil {
		utils.Error.Println(err)
		return response, err
	}
	var req interface{}
	err = json.Unmarshal(request.Service, &req)
	if err != nil {
		utils.Error.Println(err)
		return response, err
	}
	responseObj, _ := c.crdManager(req, constants.PUT)
	if responseObj.Error != "" {
		utils.Error.Println(err)
		return response, err
	}
	raw, err := json.Marshal(responseObj.Data)
	if responseObj.Error != "" {
		utils.Error.Println(err)
		return response, err
	}
	response.Service = raw
	return response, nil
}
