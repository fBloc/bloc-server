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
	BlankBody     = []byte{}
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

func doRequest(
	method string,
	headers map[string]string, remoteUrl string,
	getParam map[string]string, body []byte, respStructPointer interface{},
) (statusCode int, err error) {
	remoteUrl = buildUrl(remoteUrl, getParam)
	req, _ := http.NewRequest(method, remoteUrl, bytes.NewBuffer(body))
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	for retriedAmount := 0; retriedAmount < 3; retriedAmount++ {
		httpResp, err := httpClient.Do(req)
		if httpResp != nil {
			defer httpResp.Body.Close()
		}
		if err != nil {
			continue
		}
		body, _ := ioutil.ReadAll(httpResp.Body)
		err = json.Unmarshal(body, respStructPointer)

		if err != nil {
			continue
		}
		return httpResp.StatusCode, nil
	}
	return
}

// Get Warning: this assume resp is json data
func Get(
	headers map[string]string,
	remoteUrl string, params map[string]string, respStructPointer interface{},
) (int, error) {
	return doRequest("GET", headers, remoteUrl, params, BlankBody, respStructPointer)
}

// Delete Warning: this assume resp is json data
func Delete(
	headers map[string]string,
	remoteUrl string,
	params map[string]string,
	bodyByte []byte,
	respStructPointer interface{},
) (statusCode int, err error) {
	return doRequest("DELETE", headers, remoteUrl, params, bodyByte, respStructPointer)
}

// Patch Warning: this assume resp is json data
func Patch(
	headers map[string]string,
	remoteUrl string,
	params map[string]string,
	bodyByte []byte,
	respStructPointer interface{},
) (statusCode int, err error) {
	return doRequest("PATCH", headers, remoteUrl, params, bodyByte, respStructPointer)
}

// Post Warning: this assume req/resp is all json data
func Post(
	headers map[string]string,
	remoteUrl string,
	params map[string]string,
	bodyByte []byte,
	respIns interface{},
) (int, error) {
	return doRequest("POST", headers, remoteUrl, params, bodyByte, respIns)
}
