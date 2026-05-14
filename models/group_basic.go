package models

type GroupBasic struct {
	Uuid      string `json:"uuid"`                                                      // 唯一标识
	Name      string `json:"name"`                                                      // 分组名称
	CreatedAt int64  `gorm:"column:created_at; autoCreateTime:milli" json:"created_at"` // 创建时间，时间戳，毫秒
	UpdatedAt int64  `gorm:"column:updated_at; autoUpdateTime:milli" json:"updated_at"` // 更新时间，时间戳，毫秒
}

func (table *GroupBasic) TableName() string {
	return "group_basic"
}

// CountForName 保存时名称判重
func (table *GroupBasic) CountForName() (int64, error) {
	var cnt int64
	tx := DB.Model(table)
	if table.Uuid != "" {
		tx = tx.Where("uuid != ?", table.Uuid)
	}
	if table.Name != "" {
		tx = tx.Where("name = ?", table.Name)
	}
	return cnt, tx.Count(&cnt).Error
}

// First 获取单个分组
func (table *GroupBasic) First() error {
	tx := DB.Model(table)
	if table.Uuid != "" {
		tx = tx.Where("uuid = ?", table.Uuid)
	}
	if table.Name != "" {
		tx = tx.Where("name = ?", table.Name)
	}
	return tx.First(table).Error
}
