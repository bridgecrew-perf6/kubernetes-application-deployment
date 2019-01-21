package controllers

import (
	"github.com/gin-gonic/gin"
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
// @Description deploy services on kubernetes cluste
// @Accept  json
// @Produce  json
// @router /api/v1/kubernetes/deploy [post]
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
