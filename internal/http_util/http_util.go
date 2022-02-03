package http_util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/fBloc/bloc-server/internal/util"
)

var (
	httpClient    *http.Client
	BlankHeader   = map[string]string{}
	BlankGetParam = map[string]string{}
)

const urlPrefix = "http://"

func init() {
	t := http.DefaultTransport.(*http.Transport).Clone()
	t.MaxIdleConns = 100
	t.MaxConnsPerHost = 50
	t.MaxIdleConnsPerHost = 20
	httpClient = &http.Client{Transport: t}
}

func buildUrl(url string, getParam map[string]string) string {
	if !strings.HasPrefix(url, urlPrefix) {
		url = urlPrefix + url
	}
	var params []string
	for k, v := range getParam {
		params = append(params, fmt.Sprintf("%s=%s", k, util.UrlEncode(v)))
	}
	if len(params) > 0 {
		url = fmt.Sprintf("%s?%s", url, strings.Join(params, "&"))
	}
	return url
}

func get(
	headers map[string]string, remoteUrl string,
	getParam map[string]string, respStructPointer interface{},
) (int, error) {
	remoteUrl = buildUrl(remoteUrl, getParam)

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
	headers map[string]string,
	remoteUrl string, params map[string]string, respStructPointer interface{},
) (statusCode int, err error) {
	for retriedAmount := 0; retriedAmount < 3; retriedAmount++ {
		statusCode, err = get(headers, remoteUrl, params, respStructPointer)
		if err == nil {
			return
		}
	}
	return
}

// Delete Warning: this assume resp is json data
func Delete(
	headers map[string]string,
	remoteUrl string,
	params map[string]string,
	respStructPointer interface{},
) (statusCode int, err error) {
	remoteUrl = buildUrl(remoteUrl, params)

	req, err := http.NewRequest("DELETE", remoteUrl, nil)
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

// Post Warning: this assume req/resp is all json data
func Post(
	headers map[string]string,
	remoteUrl string,
	params map[string]string,
	bodyByte []byte,
	respIns interface{},
) (int, error) {
	remoteUrl = buildUrl(remoteUrl, params)

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
