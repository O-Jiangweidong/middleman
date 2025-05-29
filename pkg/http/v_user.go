package pkg

import (
    "fmt"
    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
    "gorm.io/gorm"
    
    "middleman/pkg/database/models"
)

func (h *ResourcesHandler) saveUser(c *gin.Context, db *gorm.DB) (err error) {
	var users []models.User
	if err = c.ShouldBindJSON(&users); err != nil {
		return err
	}

	for _, user := range users {
		var roleIds, groupIds []string
		for _, role := range user.Roles {
			roleIds = append(roleIds, role.ID)
		}
		for _, group := range user.UserGroups {
			groupIds = append(groupIds, group.ID)
		}

		var roles []models.RbacRole
		if len(roleIds) > 0 {
			db.Model(&roles).Where("id IN ?", roleIds).Find(&roles)
		}

		jmsUser := user.ToJMSUser()

		var roleBindings []models.RbacRoleBinding
		for _, role := range roles {
			roleBindings = append(roleBindings, models.RbacRoleBinding{
				ID:    uuid.New().String(),
				Scope: role.Scope, UserID: user.ID, RoleID: role.ID,
				CreatedBy: user.CreatedBy, UpdatedBy: user.UpdatedBy,
			})
		}
		user.Roles = nil

		var groups []models.UserGroup
		if len(groupIds) > 0 {
			db.Model(&groups).Where("id IN ?", groupIds).Find(&groups)
		}
		user.UserGroups = groups

		var count int64
		if err = db.Model(user).Where("id = ?", user.ID).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			if err = db.Model(user).Updates(&user).Error; err != nil {
				return err
			}
		} else {
			err = db.Transaction(func(tx *gorm.DB) error {
				if err = tx.Create(&user).Error; err != nil {
					tx.Rollback()
					return err
				}
				if err = tx.Create(&roleBindings).Error; err != nil {
					tx.Rollback()
					return err
				}
				return nil
			})
			if err != nil {
				return err
			}

			err = h.jmsClient.CreateUser(jmsUser)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (h *ResourcesHandler) saveRole(c *gin.Context, db *gorm.DB) (err error) {
	var roles []models.RbacRole
	if err = c.ShouldBindJSON(&roles); err != nil {
		return err
	}
	for _, role := range roles {
		var count int64
		if err = db.Model(role).Where("id = ?", role.ID).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			if err = db.Model(role).Updates(&role).Error; err != nil {
				return err
			}
		} else {
			if err = db.Create(&role).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

func (h *ResourcesHandler) saveUserGroup(c *gin.Context, db *gorm.DB) (err error) {
	var userGroups []models.UserGroup
	if err = c.ShouldBindJSON(&userGroups); err != nil {
		return err
	}
	for _, group := range userGroups {
		var count int64
		if err = db.Model(group).Where("id = ?", group.ID).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			if err = db.Model(group).Updates(&group).Error; err != nil {
				return err
			}
		} else {
			if err = db.Create(&group).Error; err != nil {
				return err
			}
		}
	}

	return nil
}

func (h *ResourcesHandler) getUsers(c *gin.Context, db *gorm.DB, limit, offset int) (interface{}, int64, error) {
	var err error
	var users []models.User
	queryFields := map[string]bool{
		"id":        true,
		"username":  true,
		"name":      true,
		"email":     true,
		"source":    true,
		"is_active": true,
	}
	q := db.Model(&models.User{}).Preload("Roles").Preload("UserGroups")
	for key, values := range c.Request.URL.Query() {
		if h.processedParams[key] || !queryFields[key] {
			continue
		}

		if len(values) > 0 {
			q = q.Where(fmt.Sprintf("%s = ?", key), values[len(values)-1])
		}
	}

	searchFields := []string{"username", "name", "email"}
	q = h.handleSearch(c, q, searchFields)

	var count int64
	if err = q.Count(&count).Limit(limit).Offset(offset).Find(&users).Error; err != nil {
		return nil, 0, err
	}

	type tmpUser struct {
		models.User

		IsValid     bool              `json:"is_valid"`
		Roles       []models.RbacRole `json:"-"`
		OrgRoles    []models.RbacRole `json:"org_roles"`
		SystemRoles []models.RbacRole `json:"system_roles"`
	}

	var newUsers []tmpUser
	for _, user := range users {
		var orgRoles, systemRoles []models.RbacRole
		for _, role := range user.Roles {
			if role.Scope == "system" {
				systemRoles = append(systemRoles, role)
			} else {
				orgRoles = append(orgRoles, role)
			}
		}
		newUsers = append(newUsers, tmpUser{
			User: user, IsValid: user.IsValid(),
			OrgRoles: orgRoles, SystemRoles: systemRoles,
		})
	}
	return newUsers, count, nil
}
