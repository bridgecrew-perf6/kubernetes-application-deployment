package utils

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"kubernetes-services-deployment/constants"
	"net/http"
)

func TokenInfo(token string) (map[string]string, error) {
	var str string = constants.RbacURL + constants.Rbac_Token_Info
	req, _ := http.NewRequest("GET", str, nil)
	req.Header.Add("token", token)
	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	rabc_resp := make(map[string]string)
	bytedata, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(bytedata, &rabc_resp)
	if err != nil {
		return nil, err
	}
	if resp != nil && resp.StatusCode != 200 {
		if len(rabc_resp["reason"]) > 0 {
			return nil, errors.New(rabc_resp["reason"])
		}
		return nil, errors.New("Can not connect to rbac")
	}

	if len(rabc_resp["companyId"]) > 0 && len(rabc_resp["company"]) > 0 && len(rabc_resp["username"]) > 0 {
		return rabc_resp, nil
	}

	return nil, errors.New("can not get data from token")
}
