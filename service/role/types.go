package role

// ListRequest 角色列表请求
type ListRequest struct {
	Name string `json:"name"` // 角色名称筛选
}

// CreateRequest 创建角色请求
type CreateRequest struct {
	Name        string   `json:"name" binding:"required"`        // 角色名称
	Description string   `json:"description"`                    // 角色描述
	Permissions []string `json:"permissions" binding:"required"` // 权限列表
}

// UpdateRequest 更新角色请求
type UpdateRequest struct {
	Uuid        string   `json:"uuid" binding:"required"`        // 角色UUID
	Name        string   `json:"name" binding:"required"`        // 角色名称
	Description string   `json:"description"`                    // 角色描述
	Permissions []string `json:"permissions" binding:"required"` // 权限列表
}

// DeleteRequest 删除角色请求
type DeleteRequest struct {
	Uuid []string `json:"uuid" binding:"required"` // 角色UUID列表
}
