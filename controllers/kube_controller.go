package controllers

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"kubernetes-services-deployment/core"
	"kubernetes-services-deployment/types"
	"kubernetes-services-deployment/utils"
	"net/http"
)

type KubeController struct {
}

func NewController() (*KubeController, error) {
	return &KubeController{}, nil
}

// @Title Get
// @Summary deploy services on kubernetes cluster
// @Description deploy services on kubernetes cluster
// @Param	body	body 	types.ServiceRequest	true	"body for services deployment"
// @Accept  json
// @Produce  json
// @router /api/v1/solution [post]
func (c *KubeController) DeployService(g *gin.Context) {
	req := types.ServiceRequest{}
	err := g.ShouldBind(&req)
	if err != nil {
		utils.Error.Println(err)
		utils.NewError(g, http.StatusBadRequest, err)
		return
	}

	err = core.StartServiceDeployment(&req)
	if err != nil {
		g.JSON(http.StatusInternalServerError, gin.H{"status": "service deployment failed", "error": err.Error()})
	} else {
		g.JSON(http.StatusOK, gin.H{"status": "service deployed successfully", "error": nil})
	}
}

// @Title Get
// @Summary deploy services on kubernetes cluster
// @Description deploy services on kubernetes cluster
// @Param	body	body 	types.ServiceRequest	true	"body for services deployment"
// @Accept  json
// @Produce  json
// @router /api/v1/solution [get]
func (c *KubeController) GetService(g *gin.Context) {
	req := types.ServiceRequest{}
	b, err := ioutil.ReadAll(g.Request.Body)
	if err != nil {
		utils.Error.Println(err)
		utils.NewError(g, http.StatusBadRequest, err)
		return
	}
	err = json.Unmarshal(b, &req)
	utils.Info.Println(req)
	if err != nil {
		utils.Error.Println(err)
		utils.NewError(g, http.StatusBadRequest, err)
		return
	}

	responses, err := core.GetServiceDeployment(&req)
	if err != nil {
		g.JSON(http.StatusInternalServerError, gin.H{"status": "failed to fetch data", "error": err.Error()})
	} else {
		g.JSON(http.StatusOK, gin.H{"service": responses, "error": nil})
	}
}

// @Title Get
// @Summary deploy services on kubernetes cluster
// @Description deploy services on kubernetes cluster
// @Param	body	body 	types.ServiceRequest	true	"body for services deployment"
// @Accept  json
// @Produce  json
// @router /api/v1/solution [delete]
func (c *KubeController) DeleteService(g *gin.Context) {
	req := types.ServiceRequest{}
	b, err := ioutil.ReadAll(g.Request.Body)
	if err != nil {
		utils.Error.Println(err)
		utils.NewError(g, http.StatusBadRequest, err)
		return
	}
	err = json.Unmarshal(b, &req)
	utils.Info.Println(req)
	if err != nil {
		utils.Error.Println(err)
		utils.NewError(g, http.StatusBadRequest, err)
		return
	}

	responses, err := core.DeleteServiceDeployment(&req)
	if err != nil {
		g.JSON(http.StatusInternalServerError, gin.H{"status": "failed to fetch data", "error": err.Error()})
	} else {
		g.JSON(http.StatusOK, gin.H{"service": responses, "error": nil})
	}
}

// @Title Get
// @Summary deploy services on kubernetes cluster
// @Description deploy services on kubernetes cluster
// @Param	body	body 	types.ServiceRequest	true	"body for services deployment"
// @Accept  json
// @Produce  json
// @router /api/v1/solution [patch]
func (c *KubeController) PatchService(g *gin.Context) {
	req := types.ServiceRequest{}
	b, err := ioutil.ReadAll(g.Request.Body)
	if err != nil {
		utils.Error.Println(err)
		utils.NewError(g, http.StatusBadRequest, err)
		return
	}
	err = json.Unmarshal(b, &req)
	utils.Info.Println(req)
	if err != nil {
		utils.Error.Println(err)
		utils.NewError(g, http.StatusBadRequest, err)
		return
	}

	responses, err := core.PatchServiceDeployment(&req)
	if err != nil {
		g.JSON(http.StatusInternalServerError, gin.H{"status": "failed to fetch data", "error": err.Error()})
	} else {
		g.JSON(http.StatusOK, gin.H{"service": responses, "error": nil})
	}
}
