package controllers

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"kubernetes-services-deployment/constants"
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

// @Tags health
// @Produce  json
// @router /health [get]
// @Success 200 "alive!"
// @failure 404 "{"error": ""}"
// @failure 500 "{"error": ""}"
func Health(context *gin.Context) {
	context.JSON(http.StatusOK, "alive!")
}

// @Tags solutions
// @Summary deploy services on kubernetes cluster
// @Param	body	body 	types.ServiceRequest	true	"body for services deployment"
// @Param token  header  string  false    "jwt token"
// @Accept  json
// @Produce  json
// @router /solution [post]
// @Success 200 "{"service": map[string]interface{},"project_id":""}"
// @failure 404 "{"error": ""}"
// @failure 500 "{"error": ""}"
func (c *KubeController) DeploySolution(g *gin.Context) {
	cpContext := new(core.Context)
	err := cpContext.ReadLoggingParameters(g)
	if err != nil {
		utils.Error.Println(err)
		g.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req := types.ServiceRequest{}
	err = g.ShouldBind(&req)
	if err != nil {
		utils.Error.Println(err)
		utils.NewError(g, http.StatusBadRequest, err)
		return
	}
	cpContext.InitializeLogger(g.Request.Host, g.Request.Method, g.Request.RequestURI, "", *req.ProjectId)

	responses, err := core.StartServiceDeployment(&req, cpContext)
	if err != nil {
		g.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "project_id": req.ProjectId})
		cpContext.SendBackendLogs(err.Error(), constants.LOGGING_LEVEL_ERROR)
	} else {
		g.JSON(http.StatusOK, gin.H{"service": responses, "project_id": req.ProjectId})
		cpContext.SendBackendLogs(responses, constants.LOGGING_LEVEL_DEBUG)
	}
}

// @Tags solutions
// @Summary deploy services on kubernetes cluster
// @Description deploy services on kubernetes cluster
// @Param	body	body 	types.ServiceRequest	true	"body for services deployment"
// @Param token  header  string  false    "token"
// @Accept  json
// @Produce  json
// @router /solution [get]
// @Success 200 "{"service": map[string]interface{},"project_id":""}"
// @failure 404 "{"error": ""}"
// @failure 500 "{"error": ""}"
func (c *KubeController) GetSolution(g *gin.Context) {
	cpContext := new(core.Context)
	err := cpContext.ReadLoggingParameters(g)
	if err != nil {
		utils.Error.Println(err)
		g.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
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
	cpContext.InitializeLogger(g.Request.Host, g.Request.Method, g.Request.RequestURI, "", *req.ProjectId)
	responses, err := core.GetServiceDeployment(cpContext, &req)

	if err != nil {
		g.JSON(http.StatusInternalServerError, gin.H{"project_id": req.ProjectId, "error": err.Error()})
		cpContext.SendBackendLogs(err.Error(), constants.LOGGING_LEVEL_ERROR)
	} else {
		//g.JSON(http.StatusOK, gin.H{"status": "service deployed successfully", "service":responses})
		g.JSON(http.StatusOK, gin.H{"service": responses, "project_id": req.ProjectId})
		cpContext.SendBackendLogs(responses, constants.LOGGING_LEVEL_DEBUG)
	}

}

// @Tags solutions
// @Summary deploy services on kubernetes cluster
// @Description deploy services on kubernetes cluster
// @Param	body	body 	types.ServiceRequest	true	"body for services deployment"
// @Param token  header  string  false    "token"
// @Accept  json
// @Produce  json
// @router /solution/all [get]
// @Success 200 "{"service": map[string]interface{},"project_id":""}"
// @failure 404 "{"error": ""}"
// @failure 500 "{"error": ""}"
func (c *KubeController) ListSolution(g *gin.Context) {
	cpContext := new(core.Context)
	err := cpContext.ReadLoggingParameters(g)
	if err != nil {
		utils.Error.Println(err)
		g.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
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
	cpContext.InitializeLogger(g.Request.Host, g.Request.Method, g.Request.RequestURI, "", *req.ProjectId)
	responses, err := core.ListServiceDeployment(cpContext, &req)

	if err != nil {
		g.JSON(http.StatusInternalServerError, gin.H{"project_id": req.ProjectId, "error": err.Error()})
		cpContext.SendBackendLogs(err.Error(), constants.LOGGING_LEVEL_ERROR)
	} else {
		//g.JSON(http.StatusOK, gin.H{"status": "service deployed successfully", "service":responses})
		g.JSON(http.StatusOK, gin.H{"service": responses, "project_id": req.ProjectId})
		cpContext.SendBackendLogs(responses, constants.LOGGING_LEVEL_DEBUG)
	}

}

// @Tags solutions
// @Summary deploy services on kubernetes cluster
// @Description deploy services on kubernetes cluster
// @Param	body	body 	types.ServiceRequest	true	"body for services deployment"
// @Param token  header  string  false    "token"
// @Accept  json
// @Produce  json
// @router /solution [delete]
// @Success 200 "{"service": map[string]interface{},"project_id":""}"
// @failure 404 "{"error": ""}"
// @failure 500 "{"error": ""}"
func (c *KubeController) DeleteSolution(g *gin.Context) {
	cpContext := new(core.Context)
	err := cpContext.ReadLoggingParameters(g)
	if err != nil {
		utils.Error.Println(err)
		g.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
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
	cpContext.InitializeLogger(g.Request.Host, g.Request.Method, g.Request.RequestURI, "", *req.ProjectId)
	responses, err := core.DeleteServiceDeployment(cpContext, &req)
	if err != nil {
		g.JSON(http.StatusInternalServerError, gin.H{"project_id": req.ProjectId, "error": err.Error()})
		cpContext.SendBackendLogs(err.Error(), constants.LOGGING_LEVEL_ERROR)
	} else {
		//g.JSON(http.StatusOK, gin.H{"status": "service deployed successfully", "service":responses})
		g.JSON(http.StatusOK, gin.H{"service": responses, "project_id": req.ProjectId})
		cpContext.SendBackendLogs(responses, constants.LOGGING_LEVEL_DEBUG)
	}
}

// @Tags solutions
// @Summary deploy services on kubernetes cluster
// @Description deploy services on kubernetes cluster
// @Param	body	body 	types.ServiceRequest	true	"body for services deployment"
// @Param token  header  string  false    "token"
// @Accept  json
// @Produce  json
// @router /solution [patch]
// @Success 200 "{"service": map[string]interface{},"project_id":""}"
// @failure 404 "{"error": ""}"
// @failure 500 "{"error": ""}"
func (c *KubeController) PatchSolution(g *gin.Context) {
	cpContext := new(core.Context)
	err := cpContext.ReadLoggingParameters(g)
	if err != nil {
		utils.Error.Println(err)
		g.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
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
	cpContext.InitializeLogger(g.Request.Host, g.Request.Method, g.Request.RequestURI, "", *req.ProjectId)
	responses, err := core.PatchServiceDeployment(cpContext, &req)
	if err != nil {
		g.JSON(http.StatusInternalServerError, gin.H{"project_id": req.ProjectId, "error": err.Error()})
		cpContext.SendBackendLogs(err.Error(), constants.LOGGING_LEVEL_ERROR)
	} else {
		//g.JSON(http.StatusOK, gin.H{"status": "service deployed successfully", "service":responses})
		g.JSON(http.StatusOK, gin.H{"service": responses, "project_id": req.ProjectId})
		cpContext.SendBackendLogs(responses, constants.LOGGING_LEVEL_DEBUG)
	}
}

// @Tags solutions
// @Summary deploy services on kubernetes cluster
// @Description deploy services on kubernetes cluster
// @Param	body	body 	types.ServiceRequest	true	"body for services deployment"
// @Param token  header  string  false    "token"
// @Accept  json
// @Produce  json
// @router /solution [put]
// @Success 200 "{"service": map[string]interface{},"project_id":""}"
// @failure 404 "{"error": ""}"
// @failure 500 "{"error": ""}"
func (c *KubeController) PutSolution(g *gin.Context) {
	cpContext := new(core.Context)
	err := cpContext.ReadLoggingParameters(g)
	if err != nil {
		utils.Error.Println(err)
		g.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
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
	cpContext.InitializeLogger(g.Request.Host, g.Request.Method, g.Request.RequestURI, "", *req.ProjectId)
	responses, err := core.PutServiceDeployment(cpContext, &req)
	if err != nil {
		g.JSON(http.StatusInternalServerError, gin.H{"project_id": req.ProjectId, "error": err.Error()})
		cpContext.SendBackendLogs(err.Error(), constants.LOGGING_LEVEL_ERROR)
	} else {
		//g.JSON(http.StatusOK, gin.H{"status": "service deployed successfully", "service":responses})
		g.JSON(http.StatusOK, gin.H{"service": responses, "project_id": req.ProjectId})
		cpContext.SendBackendLogs(responses, constants.LOGGING_LEVEL_DEBUG)
	}
}
