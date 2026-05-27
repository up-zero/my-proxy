package tag

type ListRequest struct {
	Name string `json:"name"` // 标签名称
}

type CreateRequest struct {
	Name string `json:"name" binding:"required"` // 标签名称
}

type UpdateRequest struct {
	Uuid string `json:"uuid" binding:"required"` // 标签唯一标识
	Name string `json:"name" binding:"required"` // 标签名称
}

type DeleteRequest struct {
	Uuid string `json:"uuid" binding:"required"` // 标签唯一标识
}
