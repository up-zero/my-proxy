package models

const (
	// UserLevelRoot 超级管理员
	UserLevelRoot = "root"
	// UserLevelUser 普通用户
	UserLevelUser = "user"
)

type UserBasic struct {
	Uuid      string `json:"uuid"`                                                      // 唯一标识
	Username  string `json:"username"`                                                  // 用户名
	Password  string `json:"password"`                                                  // 密码
	Level     string `json:"level"`                                                     // 等级，root，user
	CreatedAt int64  `gorm:"column:created_at; autoCreateTime:milli" json:"created_at"` // 创建时间，时间戳，毫秒
	UpdatedAt int64  `gorm:"column:updated_at; autoUpdateTime:milli" json:"updated_at"` // 修改时间，时间戳，毫秒
}

func (table *UserBasic) TableName() string {
	return "user_basic"
}

func (table *UserBasic) First() error {
	tx := DB.Model(table)
	if table.Username != "" {
		tx = tx.Where("username = ?", table.Username)
	}
	if table.Password != "" {
		tx = tx.Where("password = ?", table.Password)
	}
	if table.Level != "" {
		tx = tx.Where("level = ?", table.Level)
	}
	return tx.First(table).Error
}

// CountForSave 总数，名称判重
func (table *UserBasic) CountForSave() (int64, error) {
	var cnt int64
	tx := DB.Model(table)
	if table.Uuid != "" {
		tx = tx.Where("uuid != ?", table.Uuid)
	}
	if table.Username != "" {
		tx = tx.Where("username = ?", table.Username)
	}
	err := tx.Count(&cnt).Error
	return cnt, err
}
