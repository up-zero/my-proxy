package node

// ListRequest 节点列表请求
type ListRequest struct{}

// ListResponse 节点列表响应
type ListResponse struct {
	List []NodeItem `json:"list"`
}

// NodeItem 节点项
type NodeItem struct {
	Uuid      string `json:"uuid"`
	Name      string `json:"name"`
	Address   string `json:"address"`
	SecretKey string `json:"secret_key"`
	Enabled   bool   `json:"enabled"`
	IsLocal   bool   `json:"is_local"`
	CreatedAt int64  `json:"created_at"`
	UpdatedAt int64  `json:"updated_at"`
}

// CreateRequest 创建节点请求
type CreateRequest struct {
	Name      string `json:"name" binding:"required"`
	Address   string `json:"address" binding:"required"`
	SecretKey string `json:"secret_key" binding:"required"`
	Enabled   *bool  `json:"enabled"`
}

// UpdateRequest 更新节点请求
type UpdateRequest struct {
	Uuid      string `json:"uuid" binding:"required"`
	Name      string `json:"name" binding:"required"`
	Address   string `json:"address" binding:"required"`
	SecretKey string `json:"secret_key" binding:"required"`
	Enabled   *bool  `json:"enabled"`
}

// DeleteRequest 删除节点请求
type DeleteRequest struct {
	Uuid string `json:"uuid" binding:"required"`
}
