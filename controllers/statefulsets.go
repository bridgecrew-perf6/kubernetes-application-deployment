package controllers

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	_ "k8s.io/api/apps/v1"
	"kubernetes-services-deployment/core"
	"kubernetes-services-deployment/utils"
	"net/http"
)

// @host engine.swagger.io
// @BasePath /api/v1/

// @Summary get status of  all kubernetes services deployment
// @Description get status of all kubernetes services deployment on a Kubernetes Cluster. If you need all services status then pass namespace=""
// @Param project_id header string	true "project id"
// @Param namespace path string true "Namespace of kubernetes cluster"
// @Accept  json
// @Produce  json
// @Router /api/v1/statefulsets/{namespace} [get]
func (c *KubeController) ListStatefulSetsStatus(g *gin.Context) {
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
	data, err := kubeClient.ListStatefulSets(namespace)
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

// @Summary get status of kubernetes services deployment
// @Description get status of kubernetes services deployment on a Kubernetes Cluster. If you need all services status then pass namespace=""
// @Param project_id header string	true "project id"
// @Param name path string true "Name of the kubernetes service"
// @Param namespace path string true "Namespace of the kubernetes service"
// @Accept  json
// @Produce  json
// @Success 200 {object} v1.StatefulSet
// @Router /api/v1/statefulsets/{name}/{namespace} [get]
func (c *KubeController) GetStatefulSetsStatus(g *gin.Context) {
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
	data, err := kubeClient.GetStatefulSet(namespace, name)
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

// @Summary get status of kubernetes services deployment
// @Description get status of kubernetes services deployment on a Kubernetes Cluster. If you need all services status then pass namespace=""
// @Param project_id header string	true "project id"
// @Param name path string true "Name of the kubernetes service"
// @Param namespace path string true "Namespace of the kubernetes service"
// @Accept  json
// @Produce  json
// @Router /api/v1/statefulsets/{name}/{namespace} [delete]
func (c *KubeController) DeleteStatefulSetsStatus(g *gin.Context) {
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
	err = kubeClient.DeleteStatefulSet(namespace, name)
	if err != nil {
		g.JSON(http.StatusInternalServerError, gin.H{"data": "", "Error": err.Error()})
		return
	}
	g.JSON(http.StatusOK, gin.H{"error": ""})
}
