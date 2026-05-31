package dom

// Income 入账资金
type Income struct {
	CommonModel
	OrderId       int    `gorm:"index" form:"OrderId" validate:"required"`                      // 关联订单
	Amount        int    `form:"Amount" validate:"required,min=1"`                              // 入账金额（分）
	PaymentSource string `gorm:"size:32" form:"PaymentSource" validate:"required"`              // 支付来源：alipay / wechat / unionpay / ...
	Status        int    `gorm:"size:8;default:0" form:"Status" query:"Status"`                 // 状态
	TransactionId string `gorm:"size:64;index" form:"TransactionId"`                            // 微信支付交易号
	NotifyData    string `gorm:"type:text" form:"NotifyData"`                                   // 支付回调原始数据（JSON）
}

func (*Income) TableName() string { return "income" }
