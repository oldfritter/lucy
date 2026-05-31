package model

import (
	"gorm.io/gorm"

	"github.com/oldfritter/lucy/dom"
	"github.com/oldfritter/lucy/util"
)

type Refund struct {
	dom.Refund

	Income *Income `gorm:"foreignKey:IncomeId" json:",omitempty"`
}

func (r *Refund) QueryParams(p map[string]string) map[string][]any {
	params := make(map[string][]any)
	if p["IncomeId"] != "" {
		params["income_id"] = []any{"=", p["IncomeId"]}
	}
	if p["Status"] != "" {
		params["status"] = []any{"=", p["Status"]}
	}
	if p["RefundNo"] != "" {
		params["refund_no"] = []any{"=", p["RefundNo"]}
	}
	if p["TransactionId"] != "" {
		params["transaction_id"] = []any{"=", p["TransactionId"]}
	}
	return params
}

func (r *Refund) GetWithPaginate(db *gorm.DB, ru *util.Response) {
	var results []Refund
	where, values := r.WhereBuild(r.QueryParams(ru.Params))
	condition := db.Model(r).Where(where, values...)
	condition.Count(&ru.Pagination.Count)
	ru.Pagination.Init()
	if err := condition.
		Order(r.TableName() + "." + ru.Pagination.Order).
		Offset((int(ru.Pagination.CurrentPage) - 1) * int(ru.Pagination.PerPage)).
		Limit(int(ru.Pagination.PerPage)).
		Find(&results).Error; err != nil {
		return
	}
	ru.Body = results
}
