package dom

// Product 商品：定义投放类型 + 数量对应的定价
type Product struct {
	CommonModel
	CaptchaType    string `gorm:"size:16" form:"CaptchaType" validate:"required"`               // 投放类型：text:4 / text:5 / text:6 / image:rotate
	CaptchaCount   int    `gorm:"default:1" form:"CaptchaCount" validate:"required,min=1"`      // 投放验证码数量
	CurrencyId     int    `form:"CurrencyId" validate:"required"`                               // 币种 ID
	Amount         int    `form:"Amount" validate:"required,min=1"`                             // 金额（分）
	PerMinuteLimit int    `gorm:"default:100" form:"PerMinuteLimit" validate:"omitempty,min=1"` // 每分钟最大验证次数
}

func (*Product) TableName() string { return "product" }

const (
	ProductCaptchaTypeText4     = "text:4"
	ProductCaptchaTypeText5     = "text:5"
	ProductCaptchaTypeText6     = "text:6"
	ProductCaptchaTypeRotateImg = "image:rotate"
)
