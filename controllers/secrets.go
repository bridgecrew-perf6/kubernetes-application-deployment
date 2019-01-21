package controllers

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"kubernetes-services-deployment/core"
	"kubernetes-services-deployment/types"
	"kubernetes-services-deployment/utils"
	"net/http"
)

// @host engine.swagger.io
// @BasePath /api/v1/

// @Summary deploy services on kubernetes cluster
// @Description deploy services on kubernetes cluster
// @Accept  json
// @Produce  json
// @router /api/v1/registry [post]
func (c *KubeController) CreateRegistrySecret(g *gin.Context) {
	req := types.RegistryRequest{}
	err := g.ShouldBind(&req)
	if err != nil {
		utils.Error.Println(err)
		utils.NewError(g, http.StatusBadRequest, err)
		return
	}
	if req.Username == "" || req.Password == "" || req.Email == "" || req.Url == "" {
		g.JSON(http.StatusInternalServerError, gin.H{"status": "service secrets creation failed.", "Error": "registry username or password or email or url is missing in request."})
		return
	}
	data, err := core.CreateDockerRegistryCredentials(&req)
	if err != nil {
		g.JSON(http.StatusInternalServerError, gin.H{"status": "service secrets credentials creation failed.", "Error": err.Error()})
		return
	}
	d, err := json.Marshal(data)
	if err != nil {
		utils.Error.Println(err)
		g.JSON(http.StatusOK, gin.H{"status": "service secrets created successfully", "error": nil, "data": ""})
		return
	}
	g.JSON(http.StatusOK, gin.H{"status": "service secrets created successfully", "error": nil, "data": string(d)})

}

// @host engine.swagger.io
// @BasePath /api/v1/

// @Summary deploy services on kubernetes cluster
// @Description deploy services on kubernetes cluster
// @Accept  json
// @Produce  json
// @router /api/v1/registry [get]
func (c *KubeController) GetRegistrySecret(g *gin.Context) {
	req := types.RegistryRequest{}
	err := g.ShouldBind(&req)
	if err != nil {
		utils.Error.Println(err)
		utils.NewError(g, http.StatusBadRequest, err)
		return
	}
	data, err := core.GetDockerRegistryCredentials(&req)
	if err != nil {
		g.JSON(http.StatusInternalServerError, gin.H{"data": "", "Error": err.Error()})
		return
	}
	d, err := json.Marshal(data)
	if err != nil {
		utils.Error.Println(err)
		g.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error", "data": ""})
		return
	}
	g.JSON(http.StatusOK, gin.H{"error": nil, "data": string(d)})
}

// @host engine.swagger.io
// @BasePath /api/v1/

// @Summary deploy services on kubernetes cluster
// @Description deploy services on kubernetes cluster
// @Accept  json
// @Produce  json
// @router /api/v1/registry [delete]
func (c *KubeController) DeleteRegistrySecret(g *gin.Context) {
	req := types.RegistryRequest{}
	err := g.ShouldBind(&req)
	if err != nil {
		utils.Error.Println(err)
		g.JSON(http.StatusBadRequest, gin.H{"status": "", "Error": err.Error()})
		return
	}
	err = core.DeleteDockerRegistryCredentials(&req)
	if err != nil {
		g.JSON(http.StatusInternalServerError, gin.H{"status": "failed to delete secrets", "Error": err.Error()})
		return
	}

	g.JSON(http.StatusOK, gin.H{"error": nil, "status": "secrets deleted successfully"})
}
