package audit

type ListRequest struct {
	Keyword string `json:"keyword"`
	Module  string `json:"module"`
	Action  string `json:"action"`
	Page    int    `json:"page" default:"1"`
	PerPage int    `json:"per_page" default:"20"`
}
