package http_util

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
)

var (
	httpClient  *http.Client
	BlankHeader = map[string]string{}
)

const urlPrefix = "http://"

func init() {
	t := http.DefaultTransport.(*http.Transport).Clone()
	t.MaxIdleConns = 100
	t.MaxConnsPerHost = 50
	t.MaxIdleConnsPerHost = 20
	httpClient = &http.Client{Transport: t}
}

func get(
	remoteUrl string, headers map[string]string, respStructPointer interface{},
) (int, error) {
	if !strings.HasPrefix(remoteUrl, urlPrefix) {
		remoteUrl = urlPrefix + remoteUrl
	}
	req, _ := http.NewRequest("GET", remoteUrl, nil)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	httpResp, err := httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer httpResp.Body.Close()
	body, _ := ioutil.ReadAll(httpResp.Body)
	err = json.Unmarshal(body, respStructPointer)
	if err != nil {
		return 0, err
	}
	return httpResp.StatusCode, nil
}

// Get Warning: this assume resp is json data
func Get(
	remoteUrl string, headers map[string]string, respStructPointer interface{},
) (statusCode int, err error) {
	for retriedAmount := 0; retriedAmount < 3; retriedAmount++ {
		statusCode, err = get(remoteUrl, headers, respStructPointer)
		if err == nil {
			return
		}
	}
	return
}

// Post Warning: this assume req/resp is all json data
func Post(
	remoteUrl string, headers map[string]string,
	bodyByte []byte, respIns interface{},
) (int, error) {
	if !strings.HasPrefix(remoteUrl, urlPrefix) {
		remoteUrl = urlPrefix + remoteUrl
	}

	req, err := http.NewRequest("POST", remoteUrl, bytes.NewBuffer(bodyByte))
	if err != nil {
		return 0, err
	}

	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return 0, err
	}

	defer resp.Body.Close()
	respBodyByte, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	return resp.StatusCode, json.Unmarshal(respBodyByte, respIns)
}
