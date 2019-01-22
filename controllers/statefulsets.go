package controllers

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"kubernetes-services-deployment/core"
	"kubernetes-services-deployment/utils"
	"net/http"
)

// @host engine.swagger.io
// @BasePath /api/v1/

// @Summary get status of  all kubernetes services deployment
// @Description get status of all kubernetes services deployment on a Kubernetes Cluster. If you need all services status then pass namespace=""
// @Param namespace path string false "Namespace of kubernetes cluster"
// @Accept  json
// @Produce  json
// @Router /api/v1/statefulsets/{namespace} [get]
func (c *KubeController) ListStatefulSetsStatus(g *gin.Context) {
	namespace := g.Param("namespace")
	username := g.GetHeader("username")
	password := g.GetHeader("password")
	hostUrl := g.GetHeader("host_url")
	if username == "" || password == "" || hostUrl == "" {
		g.JSON(http.StatusInternalServerError, gin.H{"data": "", "Error": "username, password or host_url is missing."})
		return
	}
	data, err := core.ListStatefulSets(username, password, hostUrl, namespace)
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
// @Param name path string true "Name of the kubernetes service"
// @Param namespace path string false "Namespace of the kubernetes service"
// @Accept  json
// @Produce  json
// @Router /api/v1/statefulsets/{name}/{namespace} [get]
func (c *KubeController) GetStatefulSetsStatus(g *gin.Context) {
	namespace := g.Param("namespace")
	name := g.Param("name")
	username := g.GetHeader("username")
	password := g.GetHeader("password")
	hostUrl := g.GetHeader("host_url")
	if username == "" || password == "" || hostUrl == "" {
		g.JSON(http.StatusInternalServerError, gin.H{"data": "", "Error": "username, password or host_url is missing."})
		return
	}
	if name == "" {
		g.JSON(http.StatusInternalServerError, gin.H{"data": "", "Error": "service name is not invalid"})
		return
	}
	data, err := core.GetStatefulSet(username, password, hostUrl, namespace, name)
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
// @Param name path string true "Name of the kubernetes service"
// @Param namespace path string false "Namespace of the kubernetes service"
// @Accept  json
// @Produce  json
// @Router /api/v1/statefulsets/{name}/{namespace} [delete]
func (c *KubeController) DeleteStatefulSetsStatus(g *gin.Context) {
	namespace := g.Param("namespace")
	name := g.Param("name")
	username := g.GetHeader("username")
	password := g.GetHeader("password")
	hostUrl := g.GetHeader("host_url")
	if username == "" || password == "" || hostUrl == "" {
		g.JSON(http.StatusInternalServerError, gin.H{"data": "", "Error": "username, password or host_url is missing."})
		return
	}
	if name == "" {
		g.JSON(http.StatusInternalServerError, gin.H{"data": "", "Error": "service name is not invalid"})
		return
	}
	err := core.DeleteStatefulSet(username, password, hostUrl, namespace, name)
	if err != nil {
		g.JSON(http.StatusInternalServerError, gin.H{"data": "", "Error": err.Error()})
		return
	}
	g.JSON(http.StatusOK, gin.H{"error": ""})
}
