package dom

// Campaign 投放：定义一批验证码生成任务的参数配置
type Campaign struct {
	CommonModel
	UserId            int    `gorm:"index" form:"UserId" query:"UserId" validate:"required"`       // 关联用户
	Name              string `gorm:"size:64" form:"Name" query:"Name" validate:"required,max=64"`  // 投放名称
	CaptchaType       string `gorm:"size:16;default:text4" form:"CaptchaType" query:"CaptchaType"` // 验证码类型：text4 text5 text6 rotate
	BackgroundImages  string `gorm:"type:text" form:"BackgroundImages"`                            // 背景图片URL列表（JSON 数组）
	WordBank          string `gorm:"type:text" form:"WordBank"`                                    // 文本词库（如：好好学习，天天向上，叶公好龙，武松打虎）
	UseSystemWordBank bool   `gorm:"default:false" form:"UseSystemWordBank"`                       // 词库为空时是否回退系统字库
	CaptchaCount      int    `gorm:"default:1" form:"CaptchaCount" validate:"min=1"`               // 生成验证码数量
	Status            int    `gorm:"size:8;default:0" query:"Status" form:"Status"`                // 0-待处理 1-处理中 2-已完成 3-失败
}

func (*Campaign) TableName() string { return "campaign" }
