package notifications

import (
	"kubernetes-services-deployment/constants"
	"kubernetes-services-deployment/types"
	"kubernetes-services-deployment/utils"
)

func SendLog(msg, message_type, env_id string) (int, error) {

	var data types.LoggingRequest

	data.Id = env_id
	data.Service = constants.SERVICE_NAME
	data.Environment = "environment"
	data.Level = message_type
	data.Message = msg

	response := utils.PostNotify(constants.LoggingURL+constants.LOGGING_ENDPOINT, data)
	if response.Error != nil {
		utils.Info.Println(response.Error)
		return 400, response.Error
	}
	return response.StatusCode, response.Error

}
