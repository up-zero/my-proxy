package user

type LoginRequest struct {
	Username string `json:"username"` // 用户名
	Password string `json:"password"` // 密码
}

type LoginResponse struct {
	Token        string `json:"token"`         // token
	RefreshToken string `json:"refresh_token"` // 用于延期的 token
	Username     string `json:"username"`      // 用户名
	Level        string `json:"level"`         // 用户等级, root:超级管理员
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"` // 用于延期的 token
}

type EditPasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"` // 老密码
	NewPassword string `json:"new_password" binding:"required"` // 新密码
}
