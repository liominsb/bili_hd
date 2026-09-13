package models

import "time"

// OAuthBinding 第三方身份绑定表
//
// 为什么不把 github_id 直接加到 users 表：
// 「一个用户可以绑定多个第三方账号」是一对多关系，单独一张表更贴近真实结构，
// 以后加微信/手机号登录只需要多插一行，users 表不用动。
type OAuthBinding struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UserID    uint      `gorm:"index" json:"user_id"` // 反查「这个人绑定了哪些第三方」

	// 联合唯一索引：同一个第三方账号只能绑到一个本地用户
	// 注意存的是第三方平台的数字 id（如 GitHub 的 id），不是用户名——用户名可以改名，旧名会被释放
	Provider    string `gorm:"size:32;uniqueIndex:idx_provider_uid,priority:1" json:"provider"`
	ProviderUID string `gorm:"size:64;uniqueIndex:idx_provider_uid,priority:2" json:"provider_uid"`

	Nickname string `gorm:"size:100" json:"nickname"` // 第三方平台上的显示名，仅用于展示
}
