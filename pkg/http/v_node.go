package pkg

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"strconv"
	"strings"

	"middleman/pkg/database/models"
)

const (
	DefaultNodeKey   = "1"
	DefaultNodeValue = "New node"
)

type MetaData struct {
	ID    string `json:"id"`
	Key   string `json:"key"`
	Value string `json:"value"`
}

type TreeNodeMeta struct {
	Type string   `json:"type"`
	Data MetaData `json:"data"`
}

type TreeNode struct {
	ID       string       `json:"id"`
	Name     string       `json:"name"`
	Title    string       `json:"title"`
	PID      string       `json:"pId"`
	IsParent bool         `json:"isParent"`
	Open     bool         `json:"open"`
	Meta     TreeNodeMeta `json:"meta"`
}

func (h *ResourcesHandler) getNodes(c *gin.Context, limit, offset int) (interface{}, int64, error) {
	var err error
	var nodes []models.Node
	queryFields := map[string]bool{
		"id":         true,
		"key":        true,
		"value":      true,
		"parent_key": true,
		"full_value": true,
	}
	q := h.db.Model(&models.Node{})
	for key, values := range c.Request.URL.Query() {
		if h.processedParams[key] || !queryFields[key] {
			continue
		}

		if len(values) > 0 {
			q = q.Where(fmt.Sprintf("%s = ?", key), values[len(values)-1])
		}
	}

	searchFields := []string{"value", "full_value"}
	q = h.handleSearch(c, q, searchFields)

	var count int64
	if err = q.Count(&count).Limit(limit).Offset(offset).Find(&nodes).Error; err != nil {
		return nil, 0, err
	}
	return nodes, count, nil
}

func (h *ResourcesHandler) getChildrenNodes(c *gin.Context) (interface{}, int64, error) {
	var err error
	var nodes []models.Node

	q := h.db.Model(&models.Node{})
	searchFields := []string{"value", "full_value"}
	q = h.handleSearch(c, q, searchFields)

	if c.Query("search") == "" {
		queryKey := c.DefaultQuery("key", DefaultNodeKey)
		q = q.Where("key = ? OR parent_key = ?", queryKey, queryKey)
	}
	if err = q.Find(&nodes).Error; err != nil {
		return nil, 0, err
	}

	var newNodes []TreeNode
	for _, node := range nodes {
		newNodes = append(newNodes, TreeNode{
			ID: node.Key, Name: node.Value,
			Title: node.Value, PID: node.ParentKey,
			IsParent: true, Open: true,
			Meta: TreeNodeMeta{
				Type: "node",
				Data: MetaData{
					ID:    node.ID,
					Key:   node.Key,
					Value: node.Value,
				},
			},
		})
	}
	return newNodes, -1, nil
}

func (h *ResourcesHandler) updateNode(c *gin.Context, id string) (err error) {
	type reqNode struct {
		Value string `json:"value" binding:"required"`
	}
	var req reqNode
	if err = c.ShouldBindJSON(&req); err != nil {
		return err
	}

	if err = h.db.Model(models.Node{}).Where("id = ?", id).
		Update("value", req.Value).Error; err != nil {
		return err
	}
	go h.jmsClient.UpdateNode(id, req)
	return nil
}

func (h *ResourcesHandler) saveChildrenNode(c *gin.Context) (ids []string, err error) {
	type reqNode struct {
		ID        string `json:"id"`
		Value     string `json:"value"`
		ParentID  string `json:"parent_id" binding:"required"`
		CreatedBy string `json:"created_by"`
	}
	var nodes []reqNode
	if err = c.ShouldBindJSON(&nodes); err != nil {
		return nil, err
	}

	for _, node := range nodes {
		var pNode models.Node
		if err = h.db.Model(pNode).Where("id = ?", node.ParentID).Find(&pNode).Error; err != nil {
			return nil, err
		}

		var cNodes []models.Node
		if err = h.db.Model(pNode).Where("parent_key = ?", pNode.Key).Find(&cNodes).Error; err != nil {
			return nil, err
		}

		keySerial := 0
		valueSerial := 1
		for _, n := range cNodes {
			if node.Value == n.Value {
				return nil, fmt.Errorf("node [%s] already exists", node.Value)
			}

			keyIndex := strings.LastIndex(n.Key, ":")
			if keyIndex != -1 {
				if s, err := strconv.Atoi(n.Key[keyIndex+1:]); err != nil {
					keySerial = s
				}
			}

			if n.Value == "" {
				result := strings.Split(n.Value, " ")
				if len(result) >= 2 {
					serial, err := strconv.Atoi(result[len(result)-1])
					if err != nil {
						serial = 0
					}
					if serial > valueSerial {
						valueSerial = serial
					}
				}
			}
		}

		if node.Value == "" {
			node.Value = fmt.Sprintf("%s %v", DefaultNodeValue, valueSerial+1)
		}

		cNode := models.Node{
			ID:           node.ID,
			Key:          fmt.Sprintf("%s:%d", pNode.Key, keySerial),
			Value:        node.Value,
			ChildMark:    0,
			OrgID:        models.DefaultOrgID,
			AssetsAmount: 0,
			ParentKey:    pNode.ParentKey,
			FullValue:    fmt.Sprintf("%s/%s", pNode.FullValue, node.Value),
			Comment:      "",
			CreatedBy:    node.CreatedBy,
		}
		if err = h.db.Create(&cNode).Error; err != nil {
			return nil, err
		}

		ids = append(ids, node.ID)
		go h.jmsClient.CreateChildrenNode(cNode.ToJMS(node.ParentID))
	}
	return ids, nil
}

func (h *ResourcesHandler) saveNode(c *gin.Context) (err error) {
	var nodes []models.Node
	if err = c.ShouldBindJSON(&nodes); err != nil {
		return err
	}
	for _, node := range nodes {
		var count int64
		if err = h.db.Model(node).Where("id = ?", node.ID).Count(&count).Error; err != nil {
			return err
		}

		if count > 0 {
			if err = h.db.Model(node).Omit("id").Updates(&node).Error; err != nil {
				return err
			}
		} else {
			if err = h.db.Create(&node).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

func (h *ResourcesHandler) assetNodeRelation(c *gin.Context) (err error) {
	var req struct {
		Action   string   `json:"action" binding:"required"`
		NodeID   string   `json:"node_id"`
		AssetIds []string `json:"asset_ids"`
	}
	if err = c.ShouldBindJSON(&req); err != nil {
		return err
	}

	var jmsReq struct {
		AssetIds []string `json:"assets"`
	}
	jmsReq.AssetIds = req.AssetIds

	var nodeCount int64
	err = h.db.Model(models.Node{}).Where("id = ?", req.NodeID).Limit(1).Count(&nodeCount).Error
	if err != nil {
		return err
	}
	if nodeCount != 1 {
		return fmt.Errorf("node does not exist")
	}

	var assetsCount int64
	err = h.db.Model(models.Asset{}).Where("id IN ?", jmsReq.AssetIds).Count(&assetsCount).Error
	if err != nil {
		return err
	}

	if assetsCount != int64(len(jmsReq.AssetIds)) {
		return fmt.Errorf("there are illegal ID in param assets")
	}

	if req.Action == "add" {
		var existingRelations []struct {
			AssetID string `gorm:"column:asset_id"`
		}
		if err = h.db.Raw(
			`SELECT asset_id FROM assets_asset_nodes WHERE node_id = ? AND asset_id IN (?)`,
			req.NodeID, req.AssetIds).Scan(&existingRelations).Error; err != nil {
			return err
		}

		existingMap := make(map[string]bool)
		for _, rel := range existingRelations {
			existingMap[rel.AssetID] = true
		}

		type AssetNodeRelation struct {
			NodeID  string `gorm:"column:node_id"`
			AssetID string `gorm:"column:asset_id"`
		}
		var newRelations []AssetNodeRelation
		for _, assetID := range req.AssetIds {
			if !existingMap[assetID] {
				newRelations = append(newRelations, AssetNodeRelation{
					NodeID: req.NodeID, AssetID: assetID,
				})
			}
		}

		if len(newRelations) > 0 {
			err = h.db.Table("assets_asset_nodes").CreateInBatches(&newRelations, 100).Error
			if err != nil {
				return err
			}

			go h.jmsClient.NodeWithAssetsRelation("add", req.NodeID, jmsReq)
		}

	} else if req.Action == "remove" {
		err = h.db.Exec(
			"DELETE FROM assets_asset_nodes WHERE node_id = ? AND asset_id IN (?)",
			req.NodeID, req.AssetIds).Error
		if err != nil {
			return err
		}

		go h.jmsClient.NodeWithAssetsRelation("remove", req.NodeID, jmsReq)
	}
	return nil
}
