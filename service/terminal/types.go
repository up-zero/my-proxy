package terminal

// Size 终端尺寸
type Size struct {
	Cols int `json:"cols"`
	Rows int `json:"rows"`
}

// Message WebSocket 消息结构
type Message struct {
	Type string `json:"type"` // "data", "resize", "auth"
	Data string `json:"data,omitempty"`
	Size *Size  `json:"size,omitempty"`
}
