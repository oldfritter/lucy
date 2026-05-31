package model

import (
	"gorm.io/gorm"

	"github.com/oldfritter/lucy/dom"
	"github.com/oldfritter/lucy/util"
)

type Income struct {
	dom.Income

	Order   *Order    `gorm:"foreignKey:OrderId" json:",omitempty"`
	Refunds []*Refund `gorm:"foreignKey:IncomeId" json:",omitempty"`
}

func (i *Income) QueryParams(p map[string]string) map[string][]any {
	params := make(map[string][]any)
	if p["OrderId"] != "" {
		params["order_id"] = []any{"=", p["OrderId"]}
	}
	if p["PaymentSource"] != "" {
		params["payment_source"] = []any{"=", p["PaymentSource"]}
	}
	if p["Status"] != "" {
		params["status"] = []any{"=", p["Status"]}
	}
	if p["TransactionId"] != "" {
		params["transaction_id"] = []any{"=", p["TransactionId"]}
	}
	return params
}

func (i *Income) GetWithPaginate(db *gorm.DB, r *util.Response) {
	var results []Income
	where, values := i.WhereBuild(i.QueryParams(r.Params))
	condition := db.Model(i).Where(where, values...)
	condition.Count(&r.Pagination.Count)
	r.Pagination.Init()
	if err := condition.
		Order(i.TableName() + "." + r.Pagination.Order).
		Offset((int(r.Pagination.CurrentPage) - 1) * int(r.Pagination.PerPage)).
		Limit(int(r.Pagination.PerPage)).
		Find(&results).Error; err != nil {
		return
	}
	r.Body = results
}
