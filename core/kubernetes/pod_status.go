package kubernetes

type PodStatuses struct {
	CountRunning int      `json:"count_running"`
	CountTotal   int      `json:"count_total"`
	Name         string   `json:"name"`
	Status       []Status `json:"status"`
}

type Status struct {
	State        string   `json:"state"`
	Reason       string   `json:"reason"`
	Message      string   `json:"message"`
	HostIp       string   `json:"host_ip"`
	Name         string   `json:"name"`
	GenerateName string   `json:"-"`
	VolumeIds    []string `json:"volume_ids"`
}

/*func GetPodStatus(data PodStatusInfo) []byte {


}

type NodeIPs struct {
	InternalIP string
	ExternalIP string
}

func GetNodes(data PodStatusInfo) map[string]string {


}

func GetVolumeId(data PodStatusInfo) map[string]string {


}

*/
