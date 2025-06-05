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

	req.Header.Set("Authorization", "Token "+jms.privateKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Middleman-Version", "1.0")

	return jms.client.Do(req)
}

func (jms *JumpServer) handleCreate(url string, obj interface{}) error {
	resp, err := jms.doRequest("POST", url, obj)
	if err != nil {
		return fmt.Errorf("send request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response failed: %w", err)
	}

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("create failed，status code: %d, body: %s",
			resp.StatusCode, string(body))
	}
	return nil
}

func (jms *JumpServer) handlePatch(url string, obj interface{}) error {
	resp, err := jms.doRequest("PATCH", url, obj)
	if err != nil {
		return fmt.Errorf("send request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response failed: %w", err)
	}

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("create failed，status code: %d, body: %s",
			resp.StatusCode, string(body))
	}
	return nil
}

func (jms *JumpServer) handlePut(url string, obj interface{}) error {
	resp, err := jms.doRequest("PUT", url, obj)
	if err != nil {
		return fmt.Errorf("send request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response failed: %w", err)
	}

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("create failed，status code: %d, body: %s",
			resp.StatusCode, string(body))
	}
	return nil
}

func (jms *JumpServer) CreateUser(user models.JMSUser) error {
	url := "/api/v1/users/users/"
	return jms.handleCreate(url, user)
}

func (jms *JumpServer) CreateChildrenNode(node models.JMSNode) error {
	url := fmt.Sprintf("/api/v1/assets/nodes/%s/children/", node.ParentID)
	return jms.handleCreate(url, node)
}

func (jms *JumpServer) UpdateNode(id string, data interface{}) error {
	url := fmt.Sprintf("/api/v1/assets/nodes/%s/", id)
	return jms.handlePatch(url, data)
}

func (jms *JumpServer) CreateNode(node models.Node) error {
	url := "/api/v1/assets/nodes/?action=create"
	return jms.handleCreate(url, node)
}

func (jms *JumpServer) CreatePerm(perm models.JmsAssetPermission) error {
	url := "/api/v1/perms/asset-permissions/"
	return jms.handleCreate(url, perm)
}

func (jms *JumpServer) CreateAsset(asset interface{}) error {
	var category string
	var newAsset models.Asset
	switch v := asset.(type) {
	case models.Host:
		category = "hosts"
		newAsset = v.Asset
	default:
		return fmt.Errorf("unsupport category")
	}
	url := fmt.Sprintf("/api/v1/assets/%s/?platform=%v", category, newAsset.PlatformID)
	return jms.handleCreate(url, newAsset.ToJms())
}

func (jms *JumpServer) NodeWithAssetsRelation(action, nodeID string, data interface{}) (err error) {
	url := fmt.Sprintf("/api/v1/assets/nodes/%s/assets/%s/", nodeID, action)
	return jms.handlePut(url, data)
}

func NewJumpServer(endpoint string, privateKey string) *JumpServer {
	return &JumpServer{
		endpoint:   endpoint,
		privateKey: privateKey,
		client:     &http.Client{},
	}
}
