package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type JumpServer struct {
	endpoint   string
	privateKey string
	client     *http.Client
}

func (jms *JumpServer) doRequest(method, path string, body interface{}) (*http.Response, error) {
	url := jms.endpoint + path

	var reqBody []byte
	var err error
	if body != nil {
		reqBody, err = json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("序列化请求体失败: %w", err)
		}
	}
	req, err := http.NewRequest(method, url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Authorization", "Token "+jms.privateKey)
	req.Header.Set("Content-Type", "application/json")

	return jms.client.Do(req)
}

func (jms *JumpServer) CreateUser(user interface{}) error {
	url := "/api/v1/users/users/"
	resp, err := jms.doRequest("POST", url, user)
	if err != nil {
		return fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("创建用户请求失败，状态码: %d, 响应体: %s",
			resp.StatusCode, string(body))
	}
	return nil
}

func NewJumpServer(endpoint string, privateKey string) *JumpServer {
	return &JumpServer{
		endpoint:   endpoint,
		privateKey: privateKey,
		client:     &http.Client{},
	}
}
