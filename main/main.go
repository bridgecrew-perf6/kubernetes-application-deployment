package main

import (
	"github.com/gin-gonic/gin"
	"github.com/swaggo/gin-swagger"
	"github.com/swaggo/gin-swagger/swaggerFiles"

	"io/ioutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"kubernetes-services-deployment/constants"
	"kubernetes-services-deployment/controllers"
	"kubernetes-services-deployment/controllers/docs"
	"kubernetes-services-deployment/utils"
	"os"
)

func init() {
	utils.LoggerInit(ioutil.Discard, os.Stdout, os.Stdout, os.Stderr)
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
func main() {
	r := gin.Default()
	utils.InitFlags()
	c, _ := controllers.NewController()
	v1 := r.Group("/api/v1")
	{
		dag := v1.Group("/kubernetes")
		{
			dag.POST("deploy", c.DeployService)
		}

	}
	if constants.ServicePort == "" {
		constants.ServicePort = "8089"
	}
	docs.SwaggerInfo.Title = "Kubernetes Deployment Engine"
	docs.SwaggerInfo.Description = "Kubernetes server deployment engine."
	docs.SwaggerInfo.Version = "1.0"

	docs.SwaggerInfo.BasePath = "/api/v1"
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	r.Run(":" + constants.ServicePort)
}
func main_X() {

	client, err := kubernetes.NewForConfig(&rest.Config{Host: "https://54.237.228.34:6443", Username: "cloudplex", Password: "64bdySICej", TLSClientConfig: rest.TLSClientConfig{Insecure: true}})
	utils.Error.Println(err)
	pods, err := client.CoreV1().Pods("default").List(metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}
	for i := range pods.Items {
		utils.Info.Println(pods.Items[i].Name, pods.Items[i].Namespace)
	}

}
