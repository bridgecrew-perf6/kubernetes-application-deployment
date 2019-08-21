package controllers

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"kubernetes-services-deployment/core"
	"kubernetes-services-deployment/utils"
	"net/http"
)

// @Tags deployment
// @Summary get status of  all kubernetes services deployment
// @Description get status of all kubernetes services deployment on a Kubernetes Cluster. If you need all services status then pass namespace=""
// @Param project_id header string	true "project id"
// @Param namespace path string true "Namespace of kubernetes cluster"
// @Param token  header  string  false    "jwt token"
// @Accept  json
// @Produce  json
// @Router /deployment/{namespace} [get]
func (c *KubeController) ListDeploymentStatus(g *gin.Context) {
	namespace := g.Param("namespace")
	projectId := g.GetHeader("project_id")

	if projectId == "" {
		g.JSON(http.StatusInternalServerError, gin.H{"data": "", "Error": "project_id is missing in request"})
		return
	}
	cpContext := new(core.Context)
	err := cpContext.ReadLoggingParameters(g)
	if err != nil {
		utils.Error.Println(err)
		g.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	cpContext.InitializeLogger(g.Request.Host, g.Request.Method, g.Request.RequestURI, "", projectId)
	kubeClient, err := core.GetKubernetesClient(cpContext, &projectId)
	if err != nil {
		g.JSON(http.StatusInternalServerError, gin.H{"data": "", "Error": err.Error()})
		return
	}
	data, err := kubeClient.ListDeployments(namespace)
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

// @Tags deployment
// @Summary get status of kubernetes services deployment
// @Description get status of kubernetes services deployment on a Kubernetes Cluster. If you need all services status then pass namespace=""
// @Param project_id header string	true "project id"
// @Param name path string true "Name of the kubernetes service"
// @Param namespace path string true "Namespace of the kubernetes service"
// @Param token  header  string  false    "jwt token"
// @Accept  json
// @Produce  json
// @Router /deployment/{name}/{namespace} [get]
func (c *KubeController) GetDeploymentStatus(g *gin.Context) {
	namespace := g.Param("namespace")
	name := g.Param("name")
	projectId := g.GetHeader("project_id")

	if projectId == "" {
		g.JSON(http.StatusInternalServerError, gin.H{"data": "", "Error": "project_id is missing in request"})
		return
	}
	if name == "" {
		g.JSON(http.StatusInternalServerError, gin.H{"data": "", "Error": "service name is not invalid"})
		return
	}
	cpContext := new(core.Context)
	err := cpContext.ReadLoggingParameters(g)
	if err != nil {
		utils.Error.Println(err)
		g.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	cpContext.InitializeLogger(g.Request.Host, g.Request.Method, g.Request.RequestURI, "", projectId)
	kubeClient, err := core.GetKubernetesClient(cpContext, &projectId)
	if err != nil {
		g.JSON(http.StatusInternalServerError, gin.H{"data": "", "Error": err.Error()})
		return
	}

	data, err := kubeClient.GetDeployment(name, namespace)
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

// @Tags deployment
// @Summary get status of kubernetes services deployment
// @Description get status of kubernetes services deployment on a Kubernetes Cluster. If you need all services status then pass namespace=""
// @Param project_id header string	true "project id"
// @Param name path string true "Name of the kubernetes service"
// @Param namespace path string true "Namespace of the kubernetes service"
// @Param token  header  string  false    "jwt token"
// @Accept  json
// @Produce  json
// @Router /statefulsets/{name}/{namespace} [delete]
func (c *KubeController) DeleteDeployment(g *gin.Context) {
	namespace := g.Param("namespace")
	name := g.Param("name")
	projectId := g.GetHeader("project_id")

	if projectId == "" {
		g.JSON(http.StatusInternalServerError, gin.H{"data": "", "Error": "project_id is missing in request"})
		return
	}
	if name == "" {
		g.JSON(http.StatusInternalServerError, gin.H{"data": "", "Error": "service name is not invalid"})
		return
	}
	cpContext := new(core.Context)
	err := cpContext.ReadLoggingParameters(g)
	if err != nil {
		utils.Error.Println(err)
		g.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	cpContext.InitializeLogger(g.Request.Host, g.Request.Method, g.Request.RequestURI, "", projectId)
	kubeClient, err := core.GetKubernetesClient(cpContext, &projectId)
	if err != nil {
		g.JSON(http.StatusInternalServerError, gin.H{"data": "", "Error": err.Error()})
		return
	}

	err = kubeClient.DeleteDeployment(name, namespace)
	if err != nil {
		g.JSON(http.StatusInternalServerError, gin.H{"data": "", "Error": err.Error()})
		return
	}
	g.JSON(http.StatusOK, gin.H{"error": ""})
}
