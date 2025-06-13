package utils

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    
    "middleman/pkg/database/models"
)

type JumpServer struct {
	endpoint   string
	privateKey string
	client     *http.Client
	retryer    *RetryManager
}

func (jms *JumpServer) getHeaders() map[string]string {
	return map[string]string{
		"Content-Type":      "application/json",
		"Authorization":     "Token " + jms.privateKey,
		"Middleman-Version": "1.0",
	}
}

func (jms *JumpServer) doRequest(method, path string, body interface{}) (*http.Response, error) {
	url := jms.endpoint + path

	var reqBody []byte
	var err error
	if body != nil {
		reqBody, err = json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("serializer body failed: %w", err)
		}
	}
	req, err := http.NewRequest(method, url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("new request failed: %w", err)
	}

	for key, value := range jms.getHeaders() {
		req.Header.Set(key, value)
	}
	return jms.client.Do(req)
}

func (jms *JumpServer) Get(url string) error {
	resp, err := jms.doRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("send request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response failed: %w", err)
	}

	if resp.StatusCode >= 300 {
		return fmt.Errorf("get failed，status code: %d, body: %s",
			resp.StatusCode, string(body))
	}
	return nil
}

func (jms *JumpServer) Post(url string, obj interface{}) {
	resp, err := jms.doRequest("POST", url, obj)
	if err != nil {
		jms.retryer.AddFailedRequest("POST", url, jms.getHeaders(), obj, err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		jms.retryer.AddFailedRequest("POST", url, jms.getHeaders(), obj, err)
		return
	}

	if resp.StatusCode != http.StatusCreated {
		err = fmt.Errorf("create failed，status code: %d, body: %s", resp.StatusCode, string(body))
		jms.retryer.AddFailedRequest("POST", url, jms.getHeaders(), obj, err)
		return
	}
}

func (jms *JumpServer) Patch(url string, obj interface{}) {
	resp, err := jms.doRequest("PATCH", url, obj)
	if err != nil {
		jms.retryer.AddFailedRequest("PATCH", url, jms.getHeaders(), obj, err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		jms.retryer.AddFailedRequest("PATCH", url, jms.getHeaders(), obj, err)
		return
	}

	if resp.StatusCode >= 300 {
		err = fmt.Errorf("create failed，status code: %d, body: %s", resp.StatusCode, string(body))
		jms.retryer.AddFailedRequest("PATCH", url, jms.getHeaders(), obj, err)
		return
	}
	return
}

func (jms *JumpServer) Put(url string, obj interface{}) {
	resp, err := jms.doRequest("PUT", url, obj)
	if err != nil {
		jms.retryer.AddFailedRequest("PUT", url, jms.getHeaders(), obj, err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		jms.retryer.AddFailedRequest("PUT", url, jms.getHeaders(), obj, err)
		return
	}

	if resp.StatusCode >= 300 {
		err = fmt.Errorf("put failed，status code: %d, body: %s", resp.StatusCode, string(body))
		jms.retryer.AddFailedRequest("PUT", url, jms.getHeaders(), obj, err)
		return
	}
}

func (jms *JumpServer) Delete(url, cacheKey string) {
	resp, err := jms.doRequest("DELETE", url, nil)
	if err != nil {
		jms.retryer.AddFailedRequest("DELETE", url, jms.getHeaders(), nil, err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		jms.retryer.AddFailedRequest("DELETE", url, jms.getHeaders(), nil, err)
		return
	}

	if resp.StatusCode >= 300 && resp.StatusCode != 404 {
		err = fmt.Errorf("delete failed，status code: %d, body: %s", resp.StatusCode, string(body))
		jms.retryer.AddFailedRequest("DELETE", url, jms.getHeaders(), nil, err)
		return
	}
	_ = GetCache().Delete(cacheKey)
}

func (jms *JumpServer) CreateUser(user models.JMSUser) {
	url := "/api/v1/users/users/"
	jms.Post(url, user)
}

func (jms *JumpServer) CreateChildrenNode(node models.JMSNode) {
	url := fmt.Sprintf("/api/v1/assets/nodes/%s/children/", node.ParentID)
	jms.Post(url, node)
}

func (jms *JumpServer) UpdateNode(id string, data interface{}) {
	url := fmt.Sprintf("/api/v1/assets/nodes/%s/", id)
	jms.Patch(url, data)
}

func (jms *JumpServer) CreateNode(node models.Node) {
	url := "/api/v1/assets/nodes/?action=create"
	jms.Post(url, node)
}

func (jms *JumpServer) CreatePerm(perm models.JmsAssetPermission) {
	url := "/api/v1/perms/asset-permissions/"
	jms.Post(url, perm)
}

func (jms *JumpServer) UpdatePerm(perm models.JmsAssetPermission) {
	url := fmt.Sprintf("/api/v1/perms/asset-permissions/%s/", perm.ID)
	jms.Put(url, perm)
}

func (jms *JumpServer) CreateAsset(asset interface{}) {
	var category string
	var newAsset models.Asset
	switch v := asset.(type) {
	case models.Host:
		category = "hosts"
		newAsset = v.Asset
	default:
		return
	}
	url := fmt.Sprintf("/api/v1/assets/%s/?platform=%v", category, newAsset.PlatformID)
	jms.Post(url, newAsset.ToJms())
}

func (jms *JumpServer) NodeWithAssetsRelation(action, nodeID string, data interface{}) {
	url := fmt.Sprintf("/api/v1/assets/nodes/%s/assets/%s/", nodeID, action)
	jms.Put(url, data)
}

func (jms *JumpServer) RemoveAsset(id, cacheKey string) {
	url := fmt.Sprintf("/api/v1/assets/assets/%s/", id)
	jms.Delete(url, cacheKey)
}

func (jms *JumpServer) DeletePerm(id, cacheKey string) {
	url := fmt.Sprintf("/api/v1/perms/asset-permissions/%s/", id)
	jms.Delete(url, cacheKey)
}

func (jms *JumpServer) UnblockUser(id string) {
	url := fmt.Sprintf("/api/v1/users/users/%s/unblock", id)
	jms.Patch(url, nil)
}

func (jms *JumpServer) ResetUserMFA(id string) {
	url := fmt.Sprintf("/api/v1/users/users/%s/unblock", id)
	_ = jms.Get(url)
}

func NewJumpServer(endpoint string, privateKey string) *JumpServer {
	return &JumpServer{
		endpoint:   endpoint,
		privateKey: privateKey,
		client:     &http.Client{},
		retryer:    GetRetryer(),
	}
}
