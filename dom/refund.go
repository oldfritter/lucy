package dom

// Refund 退款
type Refund struct {
	CommonModel
	IncomeId int `gorm:"index" form:"IncomeId" validate:"required"` // 关联入账记录
	Amount   int `form:"Amount" validate:"required,min=1"`          // 退款金额（分）
	Status   int `gorm:"size:8;default:0" form:"Status" query:"Status"` // 状态
}

func (*Refund) TableName() string { return "refund" }
