package dom

// Order 订单
type Order struct {
	CommonModel
	UserId         int    `gorm:"index" form:"UserId" validate:"required"`              // 关联用户
	CurrencyId     int    `form:"CurrencyId" validate:"required"`                       // 币种
	Amount         int    `form:"Amount" validate:"required,min=1"`                     // 原始金额（分）
	DeductedAmount int    `gorm:"default:0" form:"DeductedAmount" validate:"min=0"`     // 抵扣金额（分）
	FinalAmount    int    `form:"FinalAmount"`                                           // 最终金额 = Amount - DeductedAmount（分）
	OrderNo        string `gorm:"size:32;uniqueIndex" form:"OrderNo"`                   // 订单号
	Status         int    `gorm:"size:8;default:0" form:"Status" query:"Status"`        // 状态
	ReasonType     string `gorm:"size:32;index:idx_reason" form:"ReasonType" validate:"required"` // 多态来源类型（如 Campaign）
	ReasonId       int    `gorm:"index:idx_reason" form:"ReasonId" validate:"required"`          // 多态来源 ID
}

// 订单 / 入账资金 / 退款 共用状态常量
const (
	StatusPending       = 0 // 待付款
	StatusPaid          = 1 // 已付款
	StatusPartialRefund = 2 // 部分退款
	StatusFullRefund    = 3 // 全额退款
)

func (*Order) TableName() string { return "order" }
