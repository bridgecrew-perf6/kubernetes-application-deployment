package kubernetes

import "k8s.io/client-go/kubernetes"

type PodLauncher struct {
	kubeClient *kubernetes.Clientset
}

func NewPodLauncher(c *kubernetes.Clientset) *PodLauncher {
	this := new(PodLauncher)
	this.kubeClient = c
	return this
}

/*func (p *PodLauncher) CreateSidecarContainer(serv types.Service) *k8s.Pod {


}
func (p *PodLauncher) CreatePod(serv types.Service, index int32) k8s.Pod {

}

func (p *PodLauncher) createConfigMapVolumes(node *types.DockerService) ([]k8s.Volume, []k8s.VolumeMount) {


}



func (p *PodLauncher) createVolumes(node *types.DockerService, seq string) []k8s.Volume {


}

func (p *PodLauncher) createContainerVolumes(node *types.DockerService, seq string) []k8s.VolumeMount {


}

func (p *PodLauncher) createContainer(node *types.DockerService, seq string) k8s.Container {

}

func (p *PodLauncher) createContainerEnvVars(node *types.DockerService) []k8s.EnvVar {

}

func (p *PodLauncher) updateContainerEnvVars(env []k8s.EnvVar, newEnv map[string]string, projectId string) []k8s.EnvVar {

}

func (p *PodLauncher) createContainerPorts(portsList []types.DockerPorts, portRange types.DockerPortsRange) []k8s.ContainerPort {

}
*/
