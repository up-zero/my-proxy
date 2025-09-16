package info

type Reply struct {
	Version   string   `json:"version"`   // 版本信息
	Addresses []string `json:"addresses"` // 服务地址
	Username  string   `json:"username"`  // 用户名
	Password  string   `json:"password"`  // 密码
}
