package models

// NodeBasic 节点信息
type NodeBasic struct {
	Uuid      string `json:"uuid"`                                                      // 唯一标识
	Name      string `json:"name"`                                                      // 节点名称
	Address   string `json:"address"`                                                   // 节点地址（含端口，如 192.168.1.100:12312）
	SecretKey string `json:"secret_key"`                                                // 节点密钥（对应目标节点的 JwtKey）
	Enabled   bool   `json:"enabled"`                                                   // 启用状态
	IsLocal   bool   `json:"is_local"`                                                  // 是否为本地节点
	CreatedAt int64  `gorm:"column:created_at; autoCreateTime:milli" json:"created_at"` // 创建时间，时间戳，毫秒
	UpdatedAt int64  `gorm:"column:updated_at; autoUpdateTime:milli" json:"updated_at"` // 更新时间，时间戳，毫秒
}

func (table *NodeBasic) TableName() string {
	return "node_basic"
}

// All 获取所有节点
func (table *NodeBasic) All() ([]*NodeBasic, error) {
	list := make([]*NodeBasic, 0)
	err := DB.Model(table).Order("is_local DESC, created_at ASC").Find(&list).Error
	return list, err
}

// First 获取单个节点
func (table *NodeBasic) First() error {
	tx := DB.Model(table)
	if table.Uuid != "" {
		tx = tx.Where("uuid = ?", table.Uuid)
	}
	return tx.First(table).Error
}

// CountForName 名称判重
func (table *NodeBasic) CountForName() (int64, error) {
	var cnt int64
	tx := DB.Model(table)
	if table.Uuid != "" {
		tx = tx.Where("uuid != ?", table.Uuid)
	}
	if table.Name != "" {
		tx = tx.Where("name = ?", table.Name)
	}
	err := tx.Count(&cnt).Error
	return cnt, err
}

// Create 新增节点
func (table *NodeBasic) Create() error {
	return DB.Create(table).Error
}

// Update 更新节点
func (table *NodeBasic) Update() error {
	return DB.Model(table).Where("uuid = ?", table.Uuid).Updates(map[string]interface{}{
		"name":       table.Name,
		"address":    table.Address,
		"secret_key": table.SecretKey,
		"enabled":    table.Enabled,
	}).Error
}

// Delete 删除节点
func (table *NodeBasic) Delete() error {
	return DB.Where("uuid = ?", table.Uuid).Delete(table).Error
}
