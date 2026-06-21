package config

// Item 单个配置项
type Item struct {
	Key          string `json:"key"`
	Value        string `json:"value"`
	DefaultValue string `json:"default_value"`
}

// GetResponse 获取系统设置响应
type GetResponse struct {
	Items []Item `json:"items"`
}

// UpdateRequest 更新系统设置请求
type UpdateRequest struct {
	Items []UpdateItem `json:"items"`
}

// UpdateItem 单个更新项
type UpdateItem struct {
	Key   string `json:"key" binding:"required"`
	Value string `json:"value"`
}
