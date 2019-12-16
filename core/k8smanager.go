package core

import (
	"context"
	pb "kubernetes-services-deployment/core/proto"
	"kubernetes-services-deployment/utils"
	"kubernetes-services-deployment/constants"
)

type server struct {
	
}

func (s *server) CreateService (ctx context.Context,request *pb.ServiceRequest)  (response *pb.ServiceResponse,err error) {
	response = new(pb.ServiceResponse)
	c, err := GetKubernetesClient(nil,&request.ProjectId)
	if err != nil {
		utils.Error.Println(err)
		return response,err
	}
	responseObj, _ := c.crdManager(request.Service, constants.POST)
	if responseObj.Error != ""{
		utils.Error.Println(err)
		return response,err
	}
	response.Service = &responseObj.Data
	return response,nil

}
func (s *server)GetService (ctx context.Context,request *pb.ServiceRequest)   (response *pb.ServiceResponse,err error)  {
	response = new(pb.ServiceResponse)
	c, err := GetKubernetesClient(nil,&request.ProjectId)
	if err != nil {
		utils.Error.Println(err)
		return response,err
	}
	responseObj, _ := c.crdManager(request.Service, constants.GET)
	if responseObj.Error != ""{
		utils.Error.Println(err)
		return response,err
	}
	response.Service = &responseObj.Data
	return response,nil
}
func (s *server)DeleteService (ctx context.Context,request *pb.ServiceRequest)  (response *pb.ServiceResponse,err error) {
	response = new(pb.ServiceResponse)
	c, err := GetKubernetesClient(nil,&request.ProjectId)
	if err != nil {
		utils.Error.Println(err)
		return response,err
	}
	responseObj, _ := c.crdManager(request.Service, constants.DELETE)
	if responseObj.Error != ""{
		utils.Error.Println(err)
		return response,err
	}
	response.Service = &responseObj.Data
	return response,nil
}
func (s *server)PatchService (ctx context.Context,request *pb.ServiceRequest)  (response *pb.ServiceResponse,err error) {
	response = new(pb.ServiceResponse)
	c, err := GetKubernetesClient(nil,&request.ProjectId)
	if err != nil {
		utils.Error.Println(err)
		return response,err
	}
	responseObj, _ := c.crdManager(request.Service, constants.PATCH)
	if responseObj.Error != ""{
		utils.Error.Println(err)
		return response,err
	}
	response.Service = &responseObj.Data
	return response,nil
}
func (s *server) PutService (ctx context.Context,request *pb.ServiceRequest)   (response *pb.ServiceResponse,err error) {
	response = new(pb.ServiceResponse)
	c, err := GetKubernetesClient(nil,&request.ProjectId)
	if err != nil {
		utils.Error.Println(err)
		return response,err
	}
	responseObj, _ := c.crdManager(request.Service, constants.PUT)
	if responseObj.Error != ""{
		utils.Error.Println(err)
		return response,err
	}
	response.Service = &responseObj.Data
	return response,nil
}