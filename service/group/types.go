package group

type ListRequest struct {
	Name string `json:"name"` // 分组名称
}

type CreateRequest struct {
	Name string `json:"name" binding:"required"` // 分组名称
}

type UpdateRequest struct {
	Uuid string `json:"uuid" binding:"required"` // 分组唯一标识
	Name string `json:"name" binding:"required"` // 分组名称
}

type DeleteRequest struct {
	Uuid string `json:"uuid" binding:"required"` // 分组唯一标识
}
