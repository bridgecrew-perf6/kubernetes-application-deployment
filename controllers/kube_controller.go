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

// @title Kubernetes Deployment Engine
// @version 1.0
// @description Kubernetes server deployment engine.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host engine.swagger.io
// @BasePath /api/v1/

// @Title Get
// @Summary deploy services on kubernetes cluster
// @Description deploy services on kubernetes cluster
// @Param body body types.ServiceRequest true "body for template content"
// @Accept  json
// @Produce  json
// @Failure 400 {object} types.HTTPError
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
