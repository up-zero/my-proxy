package dashboard

type OverviewResponse struct {
	Summary SummaryMetrics   `json:"summary"`
	System  SystemMetrics    `json:"system"`
	Charts  DashboardCharts  `json:"charts"`
	Nodes   []NodeLoadMetric `json:"nodes"`
}

type SummaryMetrics struct {
	ProxyTotal       int     `json:"proxy_total"`
	ProxyRunning     int     `json:"proxy_running"`
	ProxyStopped     int     `json:"proxy_stopped"`
	TotalConnections int64   `json:"total_connections"`
	TotalTrafficIn   int64   `json:"total_traffic_in"`
	TotalTrafficOut  int64   `json:"total_traffic_out"`
	InboundRate      float64 `json:"inbound_rate"`
	OutboundRate     float64 `json:"outbound_rate"`
	UptimeSeconds    int64   `json:"uptime_seconds"`
	UpdatedAt        int64   `json:"updated_at"`
}

type SystemMetrics struct {
	CPUPercent       float64 `json:"cpu_percent"`
	MemoryPercent    float64 `json:"memory_percent"`
	MemoryUsed       uint64  `json:"memory_used"`
	MemoryTotal      uint64  `json:"memory_total"`
	GoMemoryAlloc    uint64  `json:"go_memory_alloc"`
	Goroutines       int     `json:"goroutines"`
	SampleIntervalMs int64   `json:"sample_interval_ms"`
}

type DashboardCharts struct {
	Traffic     []TrafficPoint    `json:"traffic"`
	Connections []ConnectionPoint `json:"connections"`
	System      []SystemPoint     `json:"system"`
}

type TrafficPoint struct {
	Timestamp    int64   `json:"timestamp"`
	InboundRate  float64 `json:"inbound_rate"`
	OutboundRate float64 `json:"outbound_rate"`
}

type ConnectionPoint struct {
	Timestamp   int64 `json:"timestamp"`
	Connections int64 `json:"connections"`
}

type SystemPoint struct {
	Timestamp     int64   `json:"timestamp"`
	CPUPercent    float64 `json:"cpu_percent"`
	MemoryPercent float64 `json:"memory_percent"`
}

type NodeLoadMetric struct {
	Uuid              string  `json:"uuid"`
	Name              string  `json:"name"`
	GroupName         string  `json:"group_name"`
	Type              string  `json:"type"`
	State             string  `json:"state"`
	ListenAddress     string  `json:"listen_address"`
	ListenPort        string  `json:"listen_port"`
	TargetAddress     string  `json:"target_address"`
	TargetPort        string  `json:"target_port"`
	ActiveConnections int64   `json:"active_connections"`
	TrafficIn         int64   `json:"traffic_in"`
	TrafficOut        int64   `json:"traffic_out"`
	InboundRate       float64 `json:"inbound_rate"`
	OutboundRate      float64 `json:"outbound_rate"`
	LoadScore         float64 `json:"load_score"`
}
