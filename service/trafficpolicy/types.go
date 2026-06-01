package trafficpolicy

type ListRequest struct {
	Name      string `json:"name"`
	ScopeType string `json:"scope_type"`
	Status    string `json:"status"`
}

type SaveRequest struct {
	Uuid                string   `json:"uuid"`
	Name                string   `json:"name" binding:"required"`
	ScopeType           string   `json:"scope_type"`
	ScopeValue          string   `json:"scope_value"`
	OutboundLimit       string   `json:"outbound_limit"`
	MaxConnections      string   `json:"max_connections"`
	PeriodQuota         string   `json:"period_quota"`
	OverLimitActionList []string `json:"over_limit_action_list"`
}

type UuidRequest struct {
	Uuid string `json:"uuid" binding:"required"`
}
