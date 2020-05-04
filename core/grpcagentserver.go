package core

import (
	"bitbucket.org/cloudplex-devs/kubernetes-services-deployment/constants"
	"bitbucket.org/cloudplex-devs/kubernetes-services-deployment/core/proto"
	"bitbucket.org/cloudplex-devs/kubernetes-services-deployment/utils"
	agent_api "bitbucket.org/cloudplex-devs/woodpecker/agent-api"
	"context"
	"encoding/json"
	"errors"
	"google.golang.org/grpc/metadata"
	"io"
	v12 "k8s.io/api/core/v1"
	"time"
)

func (agent *AgentConnection) AgentCrdManager(method constants.RequestType, request *proto.ServiceRequest) (data []byte, err error) {

	md := metadata.Pairs(
		"name", "client2",
	)
	ctxWithTimeOut, _ := context.WithTimeout(context.Background(), 100*time.Second)
	agent.agentCtx = metadata.NewOutgoingContext(ctxWithTimeOut, md)
	agent.agentClient = agent_api.NewAgentServerClient(agent.connection)
	var service interface{}
	err = json.Unmarshal(request.Service, &service)
	if err != nil {
		utils.Error.Println(err)
		return data, err
	}
	name, ok := service.(map[string]interface{})["metadata"].(map[string]interface{})["name"].(string)
	if !ok {
		utils.Error.Println(errors.New("No name field exists in service data"))
		return data, err
	}

	namespace, _ := service.(map[string]interface{})["metadata"].(map[string]interface{})["namespace"].(string)

	kind, ok := service.(map[string]interface{})["kind"].(string)
	if !ok {
		return data, errors.New("No kind field exists in service data")
	}

	switch method {
	case constants.POST:
		if namespace != "" {
			_, err := agent.agentClient.ExecKubectl(agent.agentCtx, &agent_api.ExecKubectlRequest{
				Command: "kubectl",
				Args:    []string{"get", "ns", namespace},
			})
			if err != nil {
				response, err := agent.agentClient.ExecKubectl(agent.agentCtx, &agent_api.ExecKubectlRequest{
					Command: "kubectl",
					Args:    []string{"create", "ns", namespace},
				})
				if err != nil {
					utils.Error.Println(err)
					return data, err
				}
				utils.Info.Println(response.Stdout)
			}
		}

		_, err = agent.CreateFile(name, string(request.Service))
		if err != nil {
			utils.Error.Println(err)
			return data, err
		}

		kubectlStreamResp, err := agent.agentClient.ExecKubectlStream(agent.agentCtx, &agent_api.ExecKubectlRequest{
			Command: "kubectl",
			Args:    []string{"create", "-f", "~/" + name + ".json"},
		})
		if err != nil {
			utils.Error.Println(err)
			return data, err
		}
		for {
			feature, err := kubectlStreamResp.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				utils.Error.Println(err)
			}
			utils.Info.Println(feature.Stdout)
		}

		_, err = agent.DeleteFile(name, string(request.Service))
		if err != nil {
			utils.Error.Println(err)
			return data, err
		}

		kubectlResp, err := agent.agentClient.ExecKubectl(agent.agentCtx, &agent_api.ExecKubectlRequest{
			Command: "kubectl",
			Args:    []string{"get", kind, name, "-n", namespace, "-o", "json"},
		})
		if err != nil {
			utils.Error.Println(err)
			return []byte{}, err
		}
		if err != nil {
			utils.Error.Println(err)
			return data, err
		} else {
			data, _ = json.Marshal(kubectlResp.Stdout)
		}
	case constants.GET:
		kubectlResp, err := agent.agentClient.ExecKubectl(agent.agentCtx, &agent_api.ExecKubectlRequest{
			Command: "kubectl",
			Args:    []string{"get", kind, name, "-n", namespace, "-o", "json"},
		})
		if err != nil {
			utils.Error.Println(err)
			return data, err
		} else {
			data, _ = json.Marshal(kubectlResp.Stdout)
		}
	case constants.DELETE:
		kubectlResp, err := agent.agentClient.ExecKubectl(agent.agentCtx, &agent_api.ExecKubectlRequest{
			Command: "kubectl",
			Args:    []string{"delete", kind, name, "-n", namespace},
		})
		if err != nil {
			utils.Error.Println(err)
			return data, err
		} else {
			data, _ = json.Marshal(kubectlResp.Stdout)
		}
	case constants.PATCH:

		_, err = agent.CreateFile(name, string(request.Service))
		if err != nil {
			utils.Error.Println(err)
			return data, err
		}
		kubectlStreamResp, err := agent.agentClient.ExecKubectlStream(agent.agentCtx, &agent_api.ExecKubectlRequest{
			Command: "kubectl",
			Args:    []string{"apply", "-f", "~/" + name + ".json"},
		})
		if err != nil {
			utils.Error.Println(err)
			return data, err
		}
		for {
			feature, err := kubectlStreamResp.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				utils.Error.Println(err)
			}
			utils.Info.Println(feature.Stdout)
		}

		_, err = agent.DeleteFile(name, string(request.Service))
		if err != nil {
			utils.Error.Println(err)
			return data, err
		}

		kubectlResp, err := agent.agentClient.ExecKubectl(agent.agentCtx, &agent_api.ExecKubectlRequest{
			Command: "kubectl",
			Args:    []string{"get", kind, name, "-n", namespace, "-o", "json"},
		})
		if err != nil {
			utils.Error.Println(err)
			return data, err
		} else {
			data, _ = json.Marshal(kubectlResp.Stdout)
		}

	case constants.PUT:
		kubectlResp, err := agent.agentClient.ExecKubectl(agent.agentCtx, &agent_api.ExecKubectlRequest{
			Command: "kubectl",
			Args:    []string{"delete", kind, name, "-n", namespace},
		})
		if err != nil {
			utils.Error.Println(err)
		}
		utils.Info.Println(kubectlResp)

		_, err = agent.CreateFile(name, string(request.Service))
		if err != nil {
			utils.Error.Println(err)
			return data, err
		}

		kubectlStreamResp, err := agent.agentClient.ExecKubectlStream(agent.agentCtx, &agent_api.ExecKubectlRequest{
			Command: "kubectl",
			Args:    []string{"create", "-f", "~/" + name + ".json"},
		})
		if err != nil {
			utils.Error.Println(err)
			return data, err
		}
		for {
			feature, err := kubectlStreamResp.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				utils.Error.Println(err)
			}
			utils.Info.Println(feature.Stdout)
		}

		_, err = agent.DeleteFile(name, string(request.Service))
		if err != nil {
			utils.Error.Println(err)
			return data, err
		}

		kubectlResp, err = agent.agentClient.ExecKubectl(agent.agentCtx, &agent_api.ExecKubectlRequest{
			Command: "kubectl",
			Args:    []string{"get", kind, name, "-n", namespace, "-o", "json"},
		})
		if err != nil {
			utils.Error.Println(err)
			return data, err
		} else {
			data, _ = json.Marshal(kubectlResp.Stdout)
		}

	}

	return data, nil
}

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

	kubectlResp, err := agent.agentClient.ExecKubectl(agent.agentCtx, &agent_api.ExecKubectlRequest{
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
	response, err := agent.agentClient.ExecKubectl(agent.agentCtx, &agent_api.ExecKubectlRequest{
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
		namespaceResp = append(namespaceResp, namespace.Name)
	}

	return namespaceResp, nil
}
