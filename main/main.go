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
	"kubernetes-services-deployment/utils"
	"os"
)

func init() {
	utils.LoggerInit(ioutil.Discard, os.Stdout, os.Stdout, os.Stderr)
}

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
