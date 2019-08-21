package core

import (
	"encoding/json"
	"errors"
	"gopkg.in/resty.v1"
	"kubernetes-services-deployment/constants"
	"kubernetes-services-deployment/types"
	"kubernetes-services-deployment/utils"
	"strconv"
	"strings"
)

func GetClusterMaster(c *Context, projectId, cloudProvider, profileId string) (string, string, error) {
	authorization := ""

	notification := strings.Replace(constants.CLUSTER_GET_ENDPOINT, "{cloud_provider}", strings.ToLower(cloudProvider), -1)
	clusterEndpoint := constants.ClusterAPI + notification + projectId
	utils.Info.Println("endpoint:", clusterEndpoint, "authorization:", authorization)
	c.SendBackendLogs(struct {
		Endpoint string
	}{}, constants.LOGGING_LEVEL_DEBUG)
	clusterApiClient := resty.New()
	resp, err := clusterApiClient.
		R().
		SetHeader("X-Profile-Id", profileId).
		SetHeader("projectId", projectId).
		SetHeader("token", c.GetString("token")).
		Get(clusterEndpoint)
	//utils.Info.Println(string(resp.Body()))
	if err != nil {
		utils.Error.Println(err)

		return "", "", err
	}
	clusterObj := types.Cluster{}
	err = json.Unmarshal(resp.Body(), &clusterObj)
	if err != nil {
		utils.Info.Println(err)

		return "", "", err
	}
	//utils.Info.Println(clusterObj)
	publicIp, PrivateIp := GetMasterIP(clusterObj)
	utils.Info.Println(publicIp, PrivateIp)

	c.SendBackendLogs(struct {
		PublicIp  string
		PrivateIp string
	}{
		PublicIp:  publicIp,
		PrivateIp: PrivateIp,
	}, constants.LOGGING_LEVEL_DEBUG)
	return publicIp, PrivateIp, nil

}

func GetAWSCredentials(data interface{}) (cred types.AWSCredentials, err error) {
	raw, err := json.Marshal(data)
	if err != nil {
		utils.Error.Println(err)
		return cred, err
	}
	err = json.Unmarshal(raw, &cred)
	if err != nil {
		utils.Error.Println(err)
		return cred, err
	}
	return cred, nil
}
func GetAzureCredentials(data interface{}) (cred types.AzureCredentials, err error) {
	raw, err := json.Marshal(data)
	if err != nil {
		utils.Error.Println(err)
		return cred, err
	}
	err = json.Unmarshal(raw, &cred)
	if err != nil {
		utils.Error.Println(err)
		return cred, err
	}
	return cred, nil
}
func GetMasterIP(cluster types.Cluster) (string, string) {
	for _, nodePool := range cluster.NodePools {
		nodes := nodePool.Nodes
		for _, node := range nodes {
			if nodePool.Role == "master" {
				node.Role = nodePool.Role
				return node.PublicIP, node.PrivateIP
			}

		}

	}
	return "", ""
}
func GetKubernetesCredentials(c *Context, envId string) (types.Credentials, error) {
	endpoint := constants.KubernetesEngineURL + strings.Replace(constants.KUBERNETES_GET_CREDENTIALS_ENDPOINT, "{envId}", envId, -1)
	c.SendBackendLogs(endpoint, constants.LOGGING_LEVEL_DEBUG)
	client := resty.New()
	data, err := client.SetHeader("token", c.GetString("token")).R().Get(endpoint)
	if err != nil {
		utils.Error.Println(err)
		return types.Credentials{}, err
	}
	if data.StatusCode() >= 400 {
		utils.Error.Println("Error in Kubernetes get endpoint", data.StatusCode(), data.Status(), string(data.Body()))
		return types.Credentials{}, errors.New("Error in Kubernetes get endpoint. StatusCode: " + strconv.Itoa(data.StatusCode()) + ";Status: " + data.Status() + ";Body: " + string(data.Body()))
	}

	var body types.Credentials
	err = json.Unmarshal(data.Body(), &body)

	//c.SendBackendLogs(body, constants.LOGGING_LEVEL_DEBUG)
	if err != nil {
		utils.Error.Println(err)
		return types.Credentials{}, err
	}
	return body, nil
}

func GetProject(c *Context, projectId *string) (project *types.Project, err error) {
	if projectId == nil {
		utils.Error.Println("project id is null. send valid project id in request")
		return project, errors.New("project id is null. send valid project id in request")
	}
	notification := strings.Replace(constants.ProjectEngineEndpoint, "{project_id}", *projectId, -1)
	enviornmentEndpoint := constants.EnvironmentEngineURL + notification
	utils.Info.Println(enviornmentEndpoint)
	c.SendBackendLogs(enviornmentEndpoint, constants.LOGGING_LEVEL_DEBUG)
	clusterApiClient := resty.New()
	resp, err := clusterApiClient.
		SetHeader("token", c.GetString("token")).
		R().
		Get(enviornmentEndpoint)
	if err != nil {
		utils.Error.Println(err)
		return project, err
	}
	if resp.StatusCode() >= 400 {
		utils.Error.Println("Error in Kubernetes get endpoint", resp.StatusCode(), resp.Status(), string(resp.Body()))
		return project, errors.New("Error in Kubernetes get endpoint" + strconv.Itoa(resp.StatusCode()) + resp.Status() + string(resp.Body()))
	}
	p := types.Project{}
	err = json.Unmarshal(resp.Body(), &p)
	if err != nil {
		utils.Info.Println(err)
		return project, err
	}
	if !p.Status {
		c.SendBackendLogs(p, constants.LOGGING_LEVEL_ERROR)
		return project, errors.New("internal server error while fetching environment")
	}
	//p.Data.Credentials, err = getCredentials(c, projectId, &p.Data.CredentialsProfileId, &p.Data.Cloud)
	return &p, nil
}

func getCredentials(c *Context, projectId, profileId, cloud_provider *string) (interface{}, error) {
	if projectId == nil || profileId == nil {
		utils.Error.Println("project_id/profile_id is null. send valid project_id in request")
		return nil, errors.New("project id is null. send valid project id in request")
	}
	notification := strings.Replace(constants.VaultEndpoint, "{project_id}", *projectId, -1)
	notification = strings.Replace(notification, "{cloud_provider}", *cloud_provider, -1)
	notification = strings.Replace(notification, "{profile_id}", *profileId, -1)
	vaultEndpoint := constants.VaultURL + notification
	utils.Info.Println(vaultEndpoint)
	c.SendBackendLogs(vaultEndpoint, constants.LOGGING_LEVEL_DEBUG)
	clusterApiClient := resty.New()
	resp, err := clusterApiClient.
		R().
		Get(vaultEndpoint)
	if err != nil {
		utils.Error.Println(err)
		return nil, err
	}
	if resp.StatusCode() >= 400 {
		utils.Error.Println("Error in Kubernetes get endpoint", resp.StatusCode(), resp.Status(), string(resp.Body()))
		return nil, errors.New("Error in Kubernetes get endpoint" + strconv.Itoa(resp.StatusCode()) + resp.Status() + string(resp.Body()))
	}
	p := struct {
		Credentials interface{} `json:"credentials"`
	}{}
	err = json.Unmarshal(resp.Body(), &p)
	if err != nil {
		utils.Info.Println(err)
		return nil, err
	}
	return &p.Credentials, nil
}
