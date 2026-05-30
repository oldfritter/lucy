package model

import (
	"gorm.io/gorm"

	"github.com/oldfritter/lucy/dom"
	"github.com/oldfritter/lucy/util"
)

type Order struct {
	dom.Order

	User     *User     `gorm:"foreignKey:UserId" json:",omitempty"`
	Currency *Currency `gorm:"foreignKey:CurrencyId" json:",omitempty"`
	Incomes  []*Income `gorm:"foreignKey:OrderId" json:",omitempty"`
}

func (o *Order) QueryParams(p map[string]string) map[string][]any {
	params := make(map[string][]any)
	if p["UserId"] != "" {
		params["user_id"] = []any{"=", p["UserId"]}
	}
	if p["CurrencyId"] != "" {
		params["currency_id"] = []any{"=", p["CurrencyId"]}
	}
	if p["Status"] != "" {
		params["status"] = []any{"=", p["Status"]}
	}
	if p["OrderNo"] != "" {
		params["order_no"] = []any{"=", p["OrderNo"]}
	}
	return params
}

func (o *Order) GetWithPaginate(db *gorm.DB, r *util.Response) {
	var results []Order
	where, values := o.WhereBuild(o.QueryParams(r.Params))
	condition := db.Model(o).Where(where, values...)
	condition.Count(&r.Pagination.Count)
	r.Pagination.Init()
	if err := condition.
		Order(o.TableName() + "." + r.Pagination.Order).
		Offset((int(r.Pagination.CurrentPage) - 1) * int(r.Pagination.PerPage)).
		Limit(int(r.Pagination.PerPage)).
		Find(&results).Error; err != nil {
		return
	}
	r.Body = results
}
