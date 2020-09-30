package core

import (
	"bitbucket.org/cloudplex-devs/kubernetes-services-deployment/core/proto"
	"bitbucket.org/cloudplex-devs/kubernetes-services-deployment/utils"
	agent_api "bitbucket.org/cloudplex-devs/woodpecker/agent-api"
	"context"
	"encoding/json"
	v12 "k8s.io/api/core/v1"
)

//func (agent *AgentConnection) CreateNamespace(namespace string) (response *agent_api.ExecKubectlResponse, err error) {
//	if namespace != "" {
//		_, err := agent.agentClient.ExecKubectl(agent.agentCtx, &agent_api.ExecKubectlRequest{
//			Command: "kubectl",
//			Args:    []string{"get", "ns", namespace},
//		})
//		if err != nil {
//			utils.Error.Println(err)
//			response, err = agent.agentClient.ExecKubectl(agent.agentCtx, &agent_api.ExecKubectlRequest{
//				Command: "kubectl",
//				Args:    []string{"create", "ns", namespace},
//			})
//			if err != nil {
//				utils.Error.Println(err)
//				return response, err
//			}
//		}
//	}
//
//	return response, err
//}

func (agent *AgentConnection) GetK8sResources(ctx context.Context, request *proto.KubernetesResourceRequest) (data []byte, err error) {

	kubectlResp, err := agent.AgentClient.ExecKubectl(agent.AgentCtx, &agent_api.ExecKubectlRequest{
		Command: request.Command,
		Args:    request.Args,
	})
	if err != nil {
		utils.Error.Println(err)
		return data, err
	} else {
		data = []byte(kubectlResp.Stdout[0])
	}

	return data, err
}

func (agent *AgentConnection) GetAllNameSpaces() ([]string, error) {
	response, err := agent.ExecKubectlCommand(&agent_api.ExecKubectlRequest{
		Command: "kubectl",
		Args:    []string{"get", "ns", "-o", "json"},
	})
	if err != nil {
		utils.Error.Println(err)
		return []string{}, err
	}

	var namespaces v12.NamespaceList
	err = json.Unmarshal([]byte(response.Stdout[0]), &namespaces)
	if err != nil {
		utils.Error.Println(err)
		return []string{}, err
	}

	var namespaceResp []string
	for _, namespace := range namespaces.Items {
		if namespace.Name == "cloudplex" {
			continue
		}
		namespaceResp = append(namespaceResp, namespace.Name)
	}

	return namespaceResp, nil
}

func (agent *AgentConnection) AddLabel(ctx context.Context, request *proto.Namespacerequest) (string, error) {
	_, err := agent.AgentClient.ExecKubectl(agent.AgentCtx, &agent_api.ExecKubectlRequest{
		Command: "kubectl",
		Args:    []string{"label", "ns", request.Namespace, "istio-injection=enabled"},
	})
	if err != nil {
		utils.Error.Println(err)
		return "", err
	}
	return "successful", nil
}

func (agent *AgentConnection) Killingpod(ctx context.Context, request *proto.PodRequest) (string, error) {

	resp, err := agent.AgentClient.ExecKubectl(agent.AgentCtx, &agent_api.ExecKubectlRequest{
		Command: "kubectl",
		Args:    []string{"get", request.Type, request.Name, "-n", request.Namespace, "-o=jsonpath='{.spec.replicas}'"},
	})
	replicas := trimQuotes(resp.Stdout[0])
	_, err1 := agent.AgentClient.ExecKubectl(agent.AgentCtx, &agent_api.ExecKubectlRequest{
		Command: "kubectl",
		Args:    []string{"scale", request.Type, request.Name, "-n", request.Namespace, "--replicas=0"},
	})
	if err1 == nil {
		_, err2 := agent.AgentClient.ExecKubectl(agent.AgentCtx, &agent_api.ExecKubectlRequest{
			Command: "kubectl",
			Args:    []string{"scale", request.Type, request.Name, "-n", request.Namespace, "--replicas=" + replicas},
		})
		if err2 != nil {
			utils.Error.Println(err)
			return "", err
		}
	}

	if err != nil {
		utils.Error.Println(err)
		return "", err
	}
	return "successful", nil
}

func trimQuotes(s string) string {
	if len(s) >= 2 {
		if c := s[len(s)-1]; s[0] == c && (c == '"' || c == '\'') {
			return s[1 : len(s)-1]
		}
	}
	return s
}
