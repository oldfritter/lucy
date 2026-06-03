package dom

import (
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/oldfritter/lucy/base"
	"github.com/oldfritter/lucy/util"
)

type User struct {
	CommonModel
	DeletedAt  gorm.DeletedAt
	Password   string `gorm:"-" form:"Password" validate:"required,max=32"`                 // 用户密码原文
	Encrypted  string `gorm:"size:64"`                                                      // 密码密文
	Username   string `gorm:"size:32" form:"Username" query:"Username" validate:"required"` // 用户名，登录名
	Nickname   string `gorm:"size:32" form:"Nickname" query:"Nickname" validate:"max=32"`   // 昵称
	Phone      string `gorm:"size:16" form:"Phone" query:"Phone"`                           // 电话
	Email      string `gorm:"size:64" form:"Email" query:"Email" validate:"email"`          // Email
	Gender     string `gorm:"size:1" form:"Gender" query:"Gender" validate:"oneof=f m"`     // 性别
	Locale     string `gorm:"size:8" form:"Locale" query:"Locale"`                          // 使用语言
	InviteCode string `gorm:"size:16;uniqueIndex"`                                          // 本人邀请码
	InviterId  int    `gorm:"default:0"`                                                    // 邀请人 ID
}

func (u *User) BeforeCreate(db *gorm.DB) (err error) {
	if u.InviteCode == "" {
		u.InviteCode = util.RandStringRunes(8)
	}
	return
}

func (u *User) BeforeSave(db *gorm.DB) (err error) {
	if u.Encrypted == "" && u.Password != "" {
		u.SetEncrypted()
	}
	u.trimUsername()
	return
}

func (u *User) CompareHashAndPassword() bool {
	return bcrypt.CompareHashAndPassword([]byte(u.Encrypted), []byte(u.Password+base.DASHBOARD_DEVISE_PEPPER)) == nil
}

func (*User) TableName() string {
	return "user"
}

func (u *User) SetEncrypted() {
	b, _ := bcrypt.GenerateFromPassword([]byte(u.Password+base.DASHBOARD_DEVISE_PEPPER), bcrypt.MinCost)
	u.Encrypted = string(b)
}

func (u *User) trimUsername() {
	u.Username = strings.Trim(u.Username, " ")
}

func (u *User) GetUserId() int { return u.Id }

func (u *User) GetExpiredAt() time.Time {
	return time.Now().Add(time.Hour * 24 * time.Duration(base.JwtExpire))
}
