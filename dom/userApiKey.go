package dom

type UserApiKey struct {
	CommonModel
	UserId      int    `gorm:"index" form:"UserId" query:"UserId" validate:"required"`              // 关联用户
	CaptchaType string `gorm:"size:16;default:text4" form:"CaptchaType" query:"CaptchaType"`        // 验证码类型：text4 text5 text6 rotate
	Provider    string `gorm:"size:32" form:"Provider" query:"Provider" validate:"required,max=32"` // 模型提供商
	Name        string `gorm:"size:64" form:"Name" query:"Name" validate:"max=64"`                  // 标签名称
	Key         string `gorm:"size:256" json:"Key"`                                                 // API Key（系统自动生成，不允许输入）
	Secret      string `gorm:"size:256" json:"Secret"`                                              // API Secret（系统自动生成，不允许输入）
	IsActive    bool   `gorm:"default:true" form:"IsActive" query:"IsActive"`                       // 是否启用
}

func (*UserApiKey) TableName() string { return "user_api_key" }
