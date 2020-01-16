package agent_api

import (
	"errors"
	"gopkg.in/resty.v1"
)

func httpCaller(in *HttpRequest) ([]byte, int, error) {

	httpClient := resty.New()

	var resp *resty.Response
	var err error

	request := httpClient.R()
	request.SetBody(in.Body)

	for _, header := range in.Headers {
		request.SetHeader(header.Key, header.Value)
	}

	switch in.RequestType {
	case resty.MethodPost:
		resp, err = request.Post(in.Url)
	case resty.MethodGet:
		resp, err = request.Get(in.Url)
	case resty.MethodDelete:
		resp, err = request.Delete(in.Url)
	case resty.MethodPut:
		resp, err = request.Put(in.Url)
	case resty.MethodPatch:
		resp, err = request.Patch(in.Url)

	}

	if err != nil {
		return []byte{}, 0, err
	}

	if resp == nil {
		return []byte{}, 0, errors.New("response is nil")
	}
	return resp.Body(), resp.StatusCode(), nil

}
