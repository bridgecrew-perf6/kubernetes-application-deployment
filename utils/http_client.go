package utils

import (
	"encoding/json"
	"github.com/pkg/errors"
	"gopkg.in/resty.v1"
)

func Get(url string, body interface{}, headers map[string]string) ([]byte, error) {
	Info.Println(url)
	req := resty.New()
	r := req.R()
	if body != nil {
		req.SetAllowGetMethodPayload(true)
		r.SetBody(body)
	}
	resp, err := r.SetHeaders(headers).Get(url)
	if err != nil {
		Error.Println(err)
		return nil, err
	}
	if resp.StatusCode() >= 400 {
		Error.Println("responseCode: ", resp.StatusCode(), "\n response Body", string(resp.Body()))
		return nil, errors.New(string(resp.Body()))
	}
	Info.Println("responseCode: ", resp.StatusCode(), "\n response Body", string(resp.Body()))
	return resp.Body(), nil
}
func Patch(url string, body interface{}, headers map[string]string) ([]byte, error) {
	Info.Println(url)
	req := resty.New()
	resp, err := req.R().SetBody(body).SetHeaders(headers).Patch(url)
	if err != nil {
		Error.Println(err)
		return nil, err
	}
	if resp.StatusCode() >= 400 {
		Error.Println("responseCode: ", resp.StatusCode(), "\n response Body", string(resp.Body()))
		return nil, errors.New(string(resp.Body()))
	}
	Info.Println("responseCode: ", resp.StatusCode(), "\n response Body", string(resp.Body()))
	return resp.Body(), nil
}

func Post(url string, data interface{}, headers map[string]string) ([]byte, error) {
	b, err1 := json.Marshal(data)
	if err1 != nil {
		Info.Println(err1)
	}
	Info.Println("notification endpoint:", url)
	Info.Println("notification payload:", string(b))
	req := resty.New()

	resp, err := req.R().SetBody(data).SetHeaders(headers).Post(url)
	if err != nil {
		Error.Println(err)
		return nil, err
	}
	Info.Println("responseCode: ", resp.StatusCode(), "\n response Body", string(resp.Body()))
	return resp.Body(), nil
}

func Put(url string, body interface{}, headers map[string]string) ([]byte, error) {
	req := resty.New()
	resp, err := req.R().SetBody(body).SetHeaders(headers).Put(url)
	if err != nil {
		Error.Println(err)
		return nil, err
	}
	if resp.StatusCode() >= 400 {
		Error.Println("responseCode: ", resp.StatusCode(), "\n response Body", string(resp.Body()))
		return nil, errors.New(string(resp.Body()))
	}
	Info.Println("responseCode: ", resp.StatusCode(), "\n response Body", string(resp.Body()))
	return resp.Body(), nil
}
func Delete(url string, body interface{}, headers map[string]string) ([]byte, error) {
	req := resty.New()
	resp, err := req.R().SetBody(body).SetHeaders(headers).Delete(url)
	if err != nil {
		Error.Println(err)
		return nil, err
	}
	if resp.StatusCode() >= 400 {
		Error.Println("responseCode: ", resp.StatusCode(), "\n response Body", string(resp.Body()))
		return nil, errors.New(string(resp.Body()))
	}
	Info.Println("responseCode: ", resp.StatusCode(), "\n response Body", string(resp.Body()))
	return resp.Body(), nil
}
