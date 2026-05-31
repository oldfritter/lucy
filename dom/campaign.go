package dom

// Campaign 投放：定义一批验证码生成任务的参数配置
type Campaign struct {
	CommonModel
	UserId            int    `gorm:"index" form:"UserId" query:"UserId" validate:"required"`       // 关联用户 ID
	Name              string `gorm:"size:64" form:"Name" query:"Name" validate:"required,max=64"`  // 投放名称
	CaptchaType       string `gorm:"size:16;default:text4" form:"CaptchaType" query:"CaptchaType"` // 验证码类型：text4 / text5 / text6 / rotate
	BackgroundImages  string `gorm:"type:text" form:"BackgroundImages"`                            // 背景图片 URL 列表（JSON 数组）
	WordBank          string `gorm:"type:text" form:"WordBank"`                                    // 文本词库，换行或逗号分隔，如：好好学习，叶公好龙；为空时默认使用系统字库
	CaptchaCount      int    `gorm:"default:1" form:"CaptchaCount" validate:"min=1"`               // 用户投放为生成总量上限，系统投放为维持目标数
	Type              string `gorm:"size:16;default:user;index" form:"Type" query:"Type"`          // 投放类型：user / system
	Status            int    `gorm:"size:8;default:0" query:"Status" form:"Status"`                // 0-待处理 1-处理中 2-已完成 3-失败
}

func (*Campaign) TableName() string { return "campaign" }

// Campaign 投放类型常量
const (
	CampaignTypeUser   = "user"   // 用户投放：达到 CaptchaCount 上限后停止
	CampaignTypeSystem = "system" // 系统循环投放：维持验证码池始终有 CaptchaCount 个待用
)

// Campaign 状态常量
const (
	CampaignStatusProcessing = 1 // 处理中
	CampaignStatusAvailable  = 2 // 已完成（可用）
	CampaignStatusFailed     = 3 // 失败
)
