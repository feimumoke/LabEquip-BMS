package entity

import "github.com/feimumoke/labequipbms/defines/constant"

const UserTabName = "user_tab"

type UserTab struct {
	ID        int64                 `gorm:"column:id;primary_key" json:"id"`
	UserID    string                `gorm:"column:user_id" json:"user_id"`
	Email     string                `gorm:"column:email" json:"email"`
	Name      string                `gorm:"column:name" json:"name"`
	Avatar    string                `gorm:"column:avatar" json:"avatar"`
	Phone     string                `gorm:"column:phone" json:"phone"`
	Status    int64                 `gorm:"column:status" json:"status"`
	Passwd    string                `gorm:"column:passwd" json:"passwd"`
	Salt      string                `gorm:"column:salt" json:"salt"`
	Gender    int64                 `gorm:"column:gender" json:"gender"`
	Introduce string                `gorm:"column:introduce" json:"introduce"`
	Role      constant.UserRoleType `gorm:"column:role" json:"role"`
	Ctime     int64                 `gorm:"column:ctime" json:"ctime"`
	Mtime     int64                 `gorm:"column:mtime" json:"mtime"`
}

type UserLoginTab struct {
	ID         int64  `gorm:"column:id;primary_key" json:"id"`
	UserID     string `gorm:"column:user_id" json:"user_id"`
	LoginTime  int64  `gorm:"column:login_time" json:"login_time"`
	ExpireTime int64  `gorm:"column:expire_time" json:"expire_time"`
	LoginIP    string `gorm:"column:login_ip" json:"login_ip"`
	LoginType  int64  `gorm:"column:login_type" json:"login_type"`
	Result     int64  `gorm:"column:result" json:"result"`
}

const LoginSessionTabTableName = "login_session_tab"

// 登录session表
type LoginSession struct {
	ID         int64  `gorm:"column:id;primary_key" json:"id"`
	UserID     string `gorm:"column:user_id" json:"user_id"`
	SessionID  string `gorm:"column:session_id" json:"session_id"`
	UserInfo   string `gorm:"column:user_info" json:"user_info"`
	ExpireTime int64  `gorm:"column:expire_time" json:"expire_time"`
	Ctime      int64  `gorm:"column:ctime" json:"ctime"`
	Mtime      int64  `gorm:"column:mtime" json:"mtime"`
}

type AccountCacheInfo struct {
	UserID    string `json:"user_id"`
	UserName  string `json:"user_name"`
	Header    string `json:"header"`
	Email     string `json:"email"`
	LoginTime int64  `json:"login_time"`
	LoginIp   string `json:"login_ip"`
	Skey      string `json:"skey"`
}
