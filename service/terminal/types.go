package terminal

// Size 终端尺寸
type Size struct {
	Cols int `json:"cols"`
	Rows int `json:"rows"`
}

// DiskInfo 单个磁盘/挂载点信息
type DiskInfo struct {
	MountPoint string `json:"mount"` // 挂载点路径
	Total      uint64 `json:"total"` // 总大小 (字节)
	Used       uint64 `json:"used"`  // 已用 (字节)
}

// MonitorData 远程监控数据
type MonitorData struct {
	CPU      float64    `json:"cpu"`       // CPU 使用率 (百分比)
	MemTotal uint64     `json:"mem_total"` // 总内存 (字节)
	MemUsed  uint64     `json:"mem_used"`  // 已用内存 (字节)
	Disks    []DiskInfo `json:"disks"`     // 多磁盘/挂载点信息
}

// Message WebSocket 消息结构
type Message struct {
	Type    string       `json:"type"` // "data", "resize", "auth", "monitor", "monitor_toggle"
	Data    string       `json:"data,omitempty"`
	Size    *Size        `json:"size,omitempty"`
	Monitor *MonitorData `json:"monitor,omitempty"`
	Enabled *bool        `json:"enabled,omitempty"` // monitor_toggle 时使用
}
