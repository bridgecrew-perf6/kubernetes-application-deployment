package controllers

import (
	"bitbucket.org/cloudplex-devs/kubernetes-services-deployment/constants"
	"bitbucket.org/cloudplex-devs/kubernetes-services-deployment/core"
	"bitbucket.org/cloudplex-devs/kubernetes-services-deployment/utils"
	"github.com/gin-gonic/gin"
	"net/http"
)

// @Summary get all namespaces
// @Description get all namespaces
// @Param project_id path string	true "project id"
// @Security Bearer
// @Accept  json
// @Produce  json
// @Router /getallnamespaces/{project_id}/ [get]
// @Success 200 "{"error": "", "namespaces": ""}"
// @failure 404 "{"error": ""}"
// @failure 500 "{"error": ""}"
func (c *KubeController) GetAllNamespaces(g *gin.Context) {
	projectid := g.Param("project_id")
	token := g.GetHeader(constants.AuthTokenKey)

	if projectid == "" {
		g.JSON(http.StatusInternalServerError, gin.H{"namespaces": "", "error": "project_id is missing in request"})
		return
	}
	if token == "" {
		g.JSON(http.StatusInternalServerError, gin.H{"namespaces": "", "error": "user token is missing"})
		return
	}
	cpContext := new(core.Context)
	err := cpContext.ReadLoggingParameters(g)
	if err != nil {
		utils.Error.Println(err)
		g.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	cpContext.InitializeLogger(g.Request.Host, g.Request.Method, g.Request.RequestURI, "", projectid)

	companyId := cpContext.GetString("company_id")
	agent, err := core.GetGrpcAgentConnection()
	if err != nil {
		g.JSON(http.StatusInternalServerError, gin.H{"namespaces": "", "error": err.Error()})
		return
	}
	agent.ProjectId = projectid
	agent.CompanyId = companyId

	namespaces, err := agent.GetAllNameSpaces()
	if err != nil {
		g.JSON(http.StatusInternalServerError, gin.H{"namespaces": "", "error": err.Error()})
		return
	}
	g.JSON(http.StatusOK, gin.H{"error": "", "namespaces": namespaces})
}
