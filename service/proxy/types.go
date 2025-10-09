package proxy

type StatusRequest struct {
	Name string `json:"name"` // 代理名称
}

type CreateRequest struct {
	Name          string `json:"name" binding:"required"` // 代理名称
	Type          string `json:"type"`                    // 代理类型
	ListenAddress string `json:"listen_address"`          // 监听地址
	ListenPort    string `json:"listen_port"`             // 监听端口
	TargetAddress string `json:"target_address"`          // 目标地址
	TargetPort    string `json:"target_port"`             // 目标端口
}

type EditRequest struct {
	Uuid          string `json:"uuid"`                    // 代理唯一标识
	Name          string `json:"name" binding:"required"` // 代理名称
	Type          string `json:"type"`                    // 代理类型
	ListenAddress string `json:"listen_address"`          // 监听地址
	ListenPort    string `json:"listen_port"`             // 监听端口
	TargetAddress string `json:"target_address"`          // 目标地址
	TargetPort    string `json:"target_port"`             // 目标端口
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
