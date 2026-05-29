package dom

type Account struct {
	CommonModel
	UserId     int `form:"UserId" query:"UserId"`
	CurrencyId int `form:"CurrencyId" query:"CurrencyId"`
	Available  int `form:"Available"`
}

func (*Account) TableName() string { return "account" }
