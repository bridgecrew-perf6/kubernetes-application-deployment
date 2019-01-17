package types

type ExternalVolume struct {
	MountPoint            string `json:"volume_mount"`
	IOps                  int    `json:"iops"`
	VolumeType            string `json:"volume_type"`
	VolumeSize            int    `json:"ebs_volumes_size"`
	Ebs_type              string `json:"ebs_type"`
	VolumeID              string `json:"volume_id"`
	Delete_on_termination bool   `json:"delete_on_termination"`
	Encryption            bool   `json:"volume_encryption"`
}
