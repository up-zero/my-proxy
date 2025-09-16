package models

const (
	// UserLevelRoot 超级管理员
	UserLevelRoot = "root"
)

type UserBasic struct {
	Username  string `json:"username"`                                                  // 用户名
	Password  string `json:"password"`                                                  // 密码
	Level     string `json:"level"`                                                     // 等级，root
	CreatedAt int64  `gorm:"column:created_at; autoCreateTime:milli" json:"created_at"` // 创建时间，时间戳，毫秒
}

func (table *UserBasic) TableName() string {
	return "user_basic"
}

func (table *UserBasic) First() error {
	tx := DB.Model(table)
	if table.Username != "" {
		tx = tx.Where("username = ?", table.Username)
	}
	if table.Level != "" {
		tx = tx.Where("level = ?", table.Level)
	}
	return tx.First(table).Error
}
