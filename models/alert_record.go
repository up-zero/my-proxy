package models

const (
	AlertLevelWarning = "WARNING"
)

type AlertRecord struct {
	Uuid       string `json:"uuid"`
	Source     string `json:"source"`      // 来源模块，例如：TRAFFIC_POLICY
	SourceUuid string `json:"source_uuid"` // 来源对象 uuid
	Level      string `json:"level"`
	Title      string `json:"title"`
	Content    string `json:"content"`
	CreatedAt  int64  `gorm:"column:created_at; autoCreateTime:milli" json:"created_at"`
	UpdatedAt  int64  `gorm:"column:updated_at; autoUpdateTime:milli" json:"updated_at"`
}

func (table *AlertRecord) TableName() string {
	return "alert_record"
}
