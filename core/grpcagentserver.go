package core

import (
	agent_api "bitbucket.org/cloudplex-devs/woodpecker/agent-api"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"io"
	"kubernetes-services-deployment/constants"
	"kubernetes-services-deployment/core/proto"
	"kubernetes-services-deployment/utils"
	"time"
)

type AgentConnection struct {
	connection  *grpc.ClientConn
	agentCtx    context.Context
	agentClient agent_api.AgentServerClient
}

func GetGrpcAgentConnection() (*AgentConnection, error) {
	conn, err := grpc.Dial(constants.WoodpeckerURL, grpc.WithInsecure())
	if err != nil {
		utils.Error.Println(err)
		return &AgentConnection{}, err
	}

	return &AgentConnection{connection: conn}, nil
}

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

func (agent *AgentConnection) CreateFile(name, data string) (response *agent_api.FileResponse, err error) {
	response, err = agent.agentClient.CreateFile(agent.agentCtx, &agent_api.CreateFileRequest{
		Name: name,
		Files: []*agent_api.File{
			{
				Name: name + ".json",
				Data: data,
				Path: "~/",
			},
		},
	})
	if err != nil {
		utils.Error.Println(err)
		return response, err
	}
	utils.Info.Println(response) //status : successfully created all file
	return response, err
}

func (agent *AgentConnection) DeleteFile(name, data string) (response *agent_api.FileResponse, err error) {
	response, err = agent.agentClient.DeleteFile(agent.agentCtx, &agent_api.CreateFileRequest{
		Name: name,
		Files: []*agent_api.File{
			{
				Name: name + ".json",
				Data: data,
				Path: "~/",
			},
		},
	})
	if err != nil {
		utils.Error.Println(err)
		return response, err
	}
	utils.Info.Println(response) //status:"successfully deleted all files"
	return response, err
}

func GetSha256(projectId, companyId *string) *string {
	sha256 := sha256.Sum256([]byte(*projectId + *companyId))
	str := fmt.Sprintf("%x", sha256)
	return &str
}
