package alert

type ListRequest struct {
	Keyword string `json:"keyword"`
	Level   string `json:"level"`
	Page    int    `json:"page" default:"1"`
	PerPage int    `json:"per_page" default:"20"`
}
