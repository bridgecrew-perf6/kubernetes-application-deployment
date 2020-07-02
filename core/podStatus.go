package core

import (
	"bitbucket.org/cloudplex-devs/kubernetes-services-deployment/types"
	agent_api "bitbucket.org/cloudplex-devs/woodpecker/agent-api"
	"encoding/json"
	"strings"
)

func (agent *AgentConnection) GetPodStatus(name, namespace string) (responseObj types.PodStatus, err error) {
	kubectlResp, err := agent.agentClient.ExecKubectl(agent.agentCtx, &agent_api.ExecKubectlRequest{
		Command: "kubectl",
		Args:    []string{"get", "pods", "-n", namespace, "-o=jsonpath={.items[*].metadata.name}"},
	})
	if kubectlResp != nil && err == nil {
		arr := strings.Split(kubectlResp.Stdout[0], " ")
		for _, value := range arr {
			if strings.Contains(value, name) {
				//kubectlResp, err := agent.agentClient.ExecKubectl(agent.agentCtx, &agent_api.ExecKubectlRequest{
				//	Command: "kubectl",
				//	Args:    []string{"get", "pod",value, "-n", namespace,"-o=jsonpath={.status.phase}"},
				//})
				//if err != nil{
				//	return responseObj, err
				//}
				//phase := kubectlResp.Stdout[0]
				kubectlResp, err := agent.agentClient.ExecKubectl(agent.agentCtx, &agent_api.ExecKubectlRequest{
					Command: "kubectl",
					Args:    []string{"get", "pod", value, "-n", namespace, "-o", "json"},
				})
				if err != nil {
					return responseObj, err
				}
				var result map[string]interface{}
				err = json.Unmarshal([]byte(kubectlResp.Stdout[0]), &result)
				if err != nil {
					return responseObj, err
				}
				bytes, err := json.Marshal(result["status"])
				if err != nil {
					return responseObj, err
				}
				err = json.Unmarshal(bytes, &responseObj)

				if strings.EqualFold(responseObj.Status, "Pending") || strings.EqualFold(responseObj.Status, "Failed") || strings.EqualFold(responseObj.Status, "Unknown") {
					break
				}
			}
			continue
		}
	}
	return responseObj, nil
}
