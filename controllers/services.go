package controllers

import (
	"bitbucket.org/cloudplex-devs/kubernetes-services-deployment/core"
	"bitbucket.org/cloudplex-devs/kubernetes-services-deployment/utils"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"net/http"
)

// @Summary get status of  all kubernetes services deployment
// @Description get status of all kubernetes services deployment on a Kubernetes Cluster. If you need all services status then pass namespace=""
// @Param project_id header string	true "project id"
// @Param namespace path string true "Namespace of kubernetes cluster"
// @Security Bearer
// @Accept  json
// @Produce  json
// @Router /kubeservice/{namespace} [get]
func (c *KubeController) ListKubernetesServices(g *gin.Context) {
	namespace := g.Param("namespace")
	projectId := g.GetHeader("project_id")

	if projectId == "" {
		g.JSON(http.StatusInternalServerError, gin.H{"data": "", "error": "project_id is missing in request"})
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
		g.JSON(http.StatusInternalServerError, gin.H{"data": "", "error": err.Error()})
		return
	}
	data, err := kubeClient.ListKubernetesServices(namespace)
	if err != nil {
		g.JSON(http.StatusInternalServerError, gin.H{"data": "", "error": err.Error()})
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
// @Security Bearer
// @Accept  json
// @Produce  json
// @Router /kubeservice/{name}/{namespace} [get]
func (c *KubeController) GetKubernetesService(g *gin.Context) {
	namespace := g.Param("namespace")
	name := g.Param("name")
	projectId := g.GetHeader("project_id")

	if projectId == "" {
		g.JSON(http.StatusInternalServerError, gin.H{"data": "", "error": "project_id is missing in request"})
		return
	}
	if name == "" {
		g.JSON(http.StatusInternalServerError, gin.H{"data": "", "error": "service name is not invalid"})
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
		g.JSON(http.StatusInternalServerError, gin.H{"data": "", "error": err.Error()})
		return
	}

	data, err := kubeClient.GetKubernetesService(namespace, name)
	if err != nil {
		g.JSON(http.StatusInternalServerError, gin.H{"data": "", "error": err.Error()})
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
// @Security Bearer
// @Accept  json
// @Produce  json
// @Router /kubeservice/{name}/{namespace} [delete]
func (c *KubeController) DeleteKubernetesService(g *gin.Context) {
	namespace := g.Param("namespace")
	name := g.Param("name")
	projectId := g.GetHeader("project_id")

	if projectId == "" {
		g.JSON(http.StatusInternalServerError, gin.H{"data": "", "error": "project_id is missing in request"})
		return
	}
	if name == "" {
		g.JSON(http.StatusInternalServerError, gin.H{"data": "", "error": "service name is not invalid"})
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
		g.JSON(http.StatusInternalServerError, gin.H{"data": "", "error": err.Error()})
		return
	}

	err = kubeClient.DeleteKubernetesService(name, namespace)
	if err != nil {
		g.JSON(http.StatusInternalServerError, gin.H{"data": "", "error": err.Error()})
		return
	}
	g.JSON(http.StatusOK, gin.H{"error": ""})
}

// @Summary get status of kubernetes services deployment
// @Description get status of kubernetes services deployment on a Kubernetes Cluster. If you need all services status then pass namespace=""
// @Param OP header boolean	true "cluster name"
// @Param name path string true "Name of the kubernetes service"
// @Param namespace path string true "Namespace of the kubernetes service"
// @Param infra_id header string true "infra_id"
// @Security Bearer
// @Accept  json
// @Produce  json
// @Router /kubeservice/{name}/{namespace}/endpoint [get]
// @Success 200 "{"error": "", "external_ip": ""}"
// @failure 404 "{"error": ""}"
// @failure 500 "{"error": ""}"
func (c *KubeController) GetKubernetesServiceExternalIp(g *gin.Context) {
	namespace := g.Param("namespace")
	name := g.Param("name")
	infraId := g.GetHeader("infra_id")
	isOP := g.GetHeader("OP")

	if infraId == "" {
		g.JSON(http.StatusInternalServerError, gin.H{"external_ip": "", "error": "project_id is missing in request"})
		return
	}
	if name == "" {
		g.JSON(http.StatusInternalServerError, gin.H{"external_ip": "", "error": "service name is not invalid"})
		return
	}
	cpContext := new(core.Context)
	err := cpContext.ReadLoggingParameters(g)
	if err != nil {
		utils.Error.Println(err)
		g.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	cpContext.InitializeLogger(g.Request.Host, g.Request.Method, g.Request.RequestURI, "", infraId)

	companyId := cpContext.GetString("company_id")
	agent, err := core.GetGrpcAgentConnection()
	if err != nil {
		g.JSON(http.StatusInternalServerError, gin.H{"external_ip": "", "error": err.Error()})
		return
	}
	agent.CompanyId = companyId
	agent.InfraId = infraId

	var data string
	if isOP == "true" {
		data, err = agent.GetOPExternalIP()
		if err != nil {
			g.JSON(http.StatusInternalServerError, gin.H{"external_ip": "", "error": err.Error()})
			return
		}
	} else {
		data, err = agent.GetKubernetesServiceExternalIp(namespace, name)
		if err != nil {
			g.JSON(http.StatusInternalServerError, gin.H{"external_ip": "", "error": err.Error()})
			return
		}
	}
	g.JSON(http.StatusOK, gin.H{"error": "", "external_ip": data})
}

// @Summary get health status of kubernetes services deployment
// @Description get status of kubernetes services deployment on a Kubernetes Cluster. If you need all services status then pass namespace=""
// @Param project_id header string	true "project id"
// @Param name path string true "Name of the kubernetes service"
// @Param namespace path string true "Namespace of the kubernetes service"
// @Param infra_id header string true "infra_id"
// @Param token  header  string  false    "jwt token"
// @Accept  json
// @Produce  json
// @Router /kubeservice/clusterhealth [get]
// @Success 200 "{"error": "", "health": ""}"
// @failure 404 "{"error": ""}"
// @failure 500 "{"error": ""}"
func (c *KubeController) GetKubernetesServiceHealth(g *gin.Context) {
	infraId := g.GetHeader("infra_id")

	if infraId == "" {
		g.JSON(http.StatusInternalServerError, gin.H{"external_ip": "", "error": "project_id is missing in request"})
		return
	}

	cpContext := new(core.Context)
	err := cpContext.ReadLoggingParameters(g)
	if err != nil {
		utils.Error.Println(err)
		g.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	cpContext.InitializeLogger(g.Request.Host, g.Request.Method, g.Request.RequestURI, "", infraId)

	companyId := cpContext.GetString("company_id")
	agent, err := core.GetGrpcAgentConnection()
	if err != nil {
		g.JSON(http.StatusInternalServerError, gin.H{"external_ip": "", "error": err.Error()})
		return
	}
	agent.InfraId = infraId
	agent.CompanyId = companyId

	data, err := agent.GetKubernetesHealth()
	if err != nil {
		g.JSON(http.StatusInternalServerError, gin.H{"health": "", "error": err.Error()})
		return
	}
	g.JSON(http.StatusOK, gin.H{"error": nil, "health": data})
}
