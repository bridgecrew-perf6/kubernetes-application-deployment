package utils

import (
	"bytes"
	"encoding/json"
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
