package utils

import (
	"bitbucket.org/cloudplex-devs/kubernetes-services-deployment/types"
	"bytes"
	"encoding/json"
	"gopkg.in/resty.v1"
	"net/http"
	"reflect"
)

func Notify_Generic(state interface{}, path string) {

	url := path
	Info.Println("Notifying front end: URL: ", url)
	b, err1 := json.Marshal(state)
	if err1 != nil {
		Info.Println(err1)
	}

	_ = json.Unmarshal(b, &state)
	b1, _ := json.Marshal(state)
	Info.Println("notification payload:\n", string(b1))

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(b))

	req.Header.Set("Content-Type", "application/json")

	//tr := &http.Transport{
	//	TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	//}
	client := &http.Client{}
	//client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		Info.Println(err)
		Info.Println(reflect.TypeOf(resp))

	} else {
		statusCode := resp.StatusCode
		Info.Printf("notification status code %d\n", statusCode)

		resp.Body.Close()

	}

}
func PostNotify(url string, data interface{}) types.ResponseData {
	req := resty.New()
	resp, err := req.R().SetBody(data).Post(url)
	if err != nil {
		Error.Println(err)
		return types.ResponseData{Error: err}
	}
	return types.ResponseData{StatusCode: resp.StatusCode(), Status: resp.Status(), Body: string(resp.Body())}
}
