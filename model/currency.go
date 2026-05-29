package model

import (
	"gorm.io/gorm"

	"github.com/oldfritter/lucy/dom"
	"github.com/oldfritter/lucy/util"
)

type Currency struct {
	dom.Currency
}

func (currency *Currency) QueryParams(p map[string]string) map[string][]any {
	params := make(map[string][]any)
	if p["Name"] != "" {
		params[currency.TableName()+".name"] = []any{"like", p["Name"]}
	}
	return params
}

func (currency *Currency) GetWithPaginate(db *gorm.DB, r *util.Response) {
	var results []*Currency
	where, values := currency.WhereBuild(currency.QueryParams(r.Params))
	condition := db.Model(currency).Where(where, values...)
	condition.Count(&r.Pagination.Count)
	r.Pagination.Init()
	if err := condition.
		Order(currency.TableName() + "." + r.Pagination.Order).
		Offset((int(r.Pagination.CurrentPage) - 1) * int(r.Pagination.PerPage)).
		Limit(int(r.Pagination.PerPage)).
		Find(&results).Error; err != nil {
		return
	}
	r.Body = results
}
