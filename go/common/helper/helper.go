package helper

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"server_golang/common/crypto"
	"server_golang/common/types"
)

// HttpsRequest 发起 HTTP/HTTPS 请求
func HttpsRequest(reqURL string, data ...string) (string, error) {
	client := &http.Client{Timeout: 20 * time.Second}

	var resp *http.Response
	var err error

	if len(data) > 0 && data[0] != "" {
		resp, err = client.Post(reqURL, "application/x-www-form-urlencoded", strings.NewReader(data[0]))
	} else {
		resp, err = client.Get(reqURL)
	}

	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

// HttpsPostJSON 发起 JSON POST 请求
// 对应 PHP httpsRequest($url, json_encode($data, JSON_UNESCAPED_UNICODE)) 的场景
func HttpsPostJSON(reqURL string, jsonBody string) (string, error) {
	client := &http.Client{Timeout: 20 * time.Second}

	resp, err := client.Post(reqURL, "application/json", strings.NewReader(jsonBody))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

// CreateLinkstringUrlencode 将参数拼接为URL编码字符串
func CreateLinkstringUrlencode(params types.Map) string {
	parts := make([]string, 0, len(params))
	for k, v := range params {
		parts = append(parts, fmt.Sprintf("%s=%s", k, url.QueryEscape(types.ToString(v))))
	}
	return strings.Join(parts, "&")
}

// AuthSign 校验请求签名（基于MD5）
func AuthSign(params types.Map) bool {
	c := params.GetStringE("c")
	m := params.GetStringE("m")
	token := params.GetStringE("token")
	signature := params.GetStringE("signature")

	authParams := make(map[string]string)
	for k, v := range params {
		if k == "c" || k == "m" || k == "signature" || k == "token" {
			continue
		}
		authParams[k] = types.ToString(v)
	}

	// 按key排序
	keys := make([]string, 0, len(authParams))
	for k := range authParams {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	str := fmt.Sprintf("/%s/%s/%s/", c, m, token)
	for _, k := range keys {
		str += authParams[k]
	}

	return crypto.MD5Str(str) == signature
}
