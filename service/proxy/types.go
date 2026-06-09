package proxy

type StatusRequest struct {
	Uuid        string   `json:"uuid"`                    // 代理唯一标识
	Name        string   `json:"name"`                    // 代理名称
	TagUuidList []string `json:"tag_uuid_list,omitempty"` // 标签唯一标识列表
	SortField   string   `json:"sort_field"`              // 排序字段：name, type, listen_address, listen_port, target_address, target_port
	SortOrder   string   `json:"sort_order"`              // 排序方向：ascend, descend
}

type CreateRequest struct {
	Name           string   `json:"name" binding:"required"` // 代理名称
	TagUuidList    []string `json:"tag_uuid_list,omitempty"` // 标签唯一标识列表
	Type           string   `json:"type"`                    // 代理类型
	ListenAddress  string   `json:"listen_address"`          // 监听地址
	ListenPort     string   `json:"listen_port"`             // 监听端口
	TargetAddress  string   `json:"target_address"`          // 目标地址
	TargetPort     string   `json:"target_port"`             // 目标端口
	Socks5Username string   `json:"socks5_username"`         // SOCKS5 认证用户名
	Socks5Password string   `json:"socks5_password"`         // SOCKS5 认证密码
}

type EditRequest struct {
	Uuid           string   `json:"uuid"`                    // 代理唯一标识
	Name           string   `json:"name" binding:"required"` // 代理名称
	TagUuidList    []string `json:"tag_uuid_list,omitempty"` // 标签唯一标识列表
	Type           string   `json:"type"`                    // 代理类型
	ListenAddress  string   `json:"listen_address"`          // 监听地址
	ListenPort     string   `json:"listen_port"`             // 监听端口
	TargetAddress  string   `json:"target_address"`          // 目标地址
	TargetPort     string   `json:"target_port"`             // 目标端口
	Socks5Username string   `json:"socks5_username"`         // SOCKS5 认证用户名
	Socks5Password string   `json:"socks5_password"`         // SOCKS5 认证密码
}

type DeleteRequest struct {
	Name string `json:"name" binding:"required"` // 代理名称
}

type BatchDeleteRequest struct {
	Uuid []string `json:"uuid" binding:"required"` // 代理唯一标识
}

type StartRequest struct {
	Name string `json:"name" binding:"required"` // 代理名称
}

type StopRequest struct {
	Name string `json:"name" binding:"required"` // 代理名称
}

type RestartRequest struct {
	Name string `json:"name" binding:"required"` // 代理名称
}

type ExportRequest struct {
	Uuid []string `json:"uuid" binding:"required"` // 代理唯一标识
}
