package dom

type Currency struct {
	CommonModel
	Name   string
	Symbol string
	Unit   string
	Base   int `gorm:"default:1"`
}

func (*Currency) TableName() string {
	return "currency"
}
