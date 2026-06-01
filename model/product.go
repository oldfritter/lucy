package model

import (
	"gorm.io/gorm"

	"github.com/oldfritter/lucy/dom"
	"github.com/oldfritter/lucy/util"
)

type Product struct {
	dom.Product

	Currency *Currency `gorm:"foreignKey:CurrencyId" json:",omitempty"`
}

func (p *Product) QueryParams(query map[string]string) map[string][]any {
	params := make(map[string][]any)
	if query["CaptchaType"] != "" {
		params["captcha_type"] = []any{"=", query["CaptchaType"]}
	}
	if query["CurrencyId"] != "" {
		params["currency_id"] = []any{"=", query["CurrencyId"]}
	}
	return params
}

func (p *Product) GetWithPaginate(db *gorm.DB, r *util.Response) {
	var results []Product
	where, values := p.WhereBuild(p.QueryParams(r.Params))
	condition := db.Model(p).Where(where, values...)
	condition.Count(&r.Pagination.Count)
	r.Pagination.Init()
	if err := condition.
		Order(p.TableName() + "." + r.Pagination.Order).
		Offset((int(r.Pagination.CurrentPage) - 1) * int(r.Pagination.PerPage)).
		Limit(int(r.Pagination.PerPage)).
		Find(&results).Error; err != nil {
		return
	}
	r.Body = results
}
