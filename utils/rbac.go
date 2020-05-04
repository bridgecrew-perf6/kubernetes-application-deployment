package utils

import (
	"bitbucket.org/cloudplex-devs/kubernetes-services-deployment/constants"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
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
	rabc_resp := make(map[string]interface{})
	bytedata, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(bytedata, &rabc_resp)
	if err != nil {
		return nil, err
	}
	if resp != nil && resp.StatusCode != 200 {
		val, ok := rabc_resp["reason"].(string)
		if ok {
			if len(val) > 0 {
				return nil, errors.New(rabc_resp["reason"].(string))
			}
			return nil, errors.New("can not connect to rbac")
		}
	}
	val, ok := rabc_resp["companyId"].(string)
	temp := make(map[string]string)
	if ok {
		temp["companyId"] = val
	} else {
		return nil, errors.New("can not get data from token")
	}
	val, ok = rabc_resp["company"].(string)
	if ok {
		temp["company"] = val
	} else {
		return nil, errors.New("can not get data from token")
	}
	val, ok = rabc_resp["username"].(string)
	if ok {
		temp["username"] = val
		return temp, nil
	}
	return nil, errors.New("can not get data from token")
}
func GetUserIDCompanyID(token string) (userID, companyID string, err error) {

	out := strings.Split(token, ".")
	if len(out) != 3 {
		return "", "", errors.New("token is invalid")
	}
	decoded, err := base64.StdEncoding.DecodeString(out[1])
	//if err != nil {
	//
	//}
	var tokenData map[string]interface{}
	err = json.Unmarshal(decoded, &tokenData)
	if err != nil {
		return "", "", err
	}
	return tokenData["username"].(string), tokenData["companyId"].(string), nil
}
