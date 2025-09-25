package models

import "github.com/up-zero/gotool/randomutil"

// initData 初始化数据
func initData() {
	// 初始化用户
	initUserData()
}

func initUserData() {
	// admin
	ub := &UserBasic{
		Username: "admin",
		Password: randomutil.RandomAlphaNumber(8),
		Level:    "root",
	}
	if err := DB.FirstOrCreate(ub, &UserBasic{Username: ub.Username}).Error; err != nil {
		panic(err)
	}
}
