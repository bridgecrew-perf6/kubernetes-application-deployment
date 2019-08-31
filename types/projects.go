package types

type Project struct {
	Status bool `json:"status"`
	Data   struct {
		Cloud                string      `json:"cloud"`
		Region               string      `json:"region"`
		CredentialsProfileId string      `json:"profile_id"`
		Credentials          interface{} `json:"credentials"`
		/*ID            string `json:"_id"`
		EnvironmentID string `json:"environment_id"`
		Name          string `json:"name"`
		Description   string `json:"description"`
		Network       string `json:"network"`
		Cluster       string `json:"cluster"`
		Solution      string `json:"solution"`
		Status        string `json:"status"`*/
	} `json:"data"`
}
