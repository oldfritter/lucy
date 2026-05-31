package dom

// Campaign 投放：定义一批验证码生成任务的参数配置
type Campaign struct {
	CommonModel
	UserId            int    `gorm:"index" form:"UserId" query:"UserId" validate:"required"`
	Name              string `gorm:"size:64" form:"Name" query:"Name" validate:"required,max=64"`
	CaptchaType       string `gorm:"size:16;default:text4" form:"CaptchaType" query:"CaptchaType"`
	BackgroundImages  string `gorm:"type:text" form:"BackgroundImages"`
	WordBank          string `gorm:"type:text" form:"WordBank"`
	UseSystemWordBank bool   `gorm:"default:false" form:"UseSystemWordBank"`
	CaptchaCount      int    `gorm:"default:1" form:"CaptchaCount" validate:"min=1"`
	Status            int    `gorm:"size:8;default:0" query:"Status" form:"Status"`
}

func (*Campaign) TableName() string { return "campaign" }

// Campaign 状态常量
const (
	CampaignStatusProcessing = 1
	CampaignStatusAvailable  = 2
	CampaignStatusFailed     = 3
)
