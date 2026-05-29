package model

import (
	"gorm.io/gorm"

	"github.com/oldfritter/lucy/dom"
	"github.com/oldfritter/lucy/util"
)

type Account struct {
	dom.Account

	User            *User                 `gorm:"foreignKey:UserId" json:",omitempty"`
	Currency        *Currency             `gorm:"foreignKey:CurrencyId" json:",omitempty"`
	AccountVersions []*dom.AccountVersion `gorm:"foreignKey:AccountId" json:",omitempty"`
}

func (account *Account) QueryParams(p map[string]string) map[string][]any {
	params := make(map[string][]any)
	if p["UserId"] != "" {
		params["user_id"] = []any{"=", p["UserId"]}
	}
	if p["CurrencyId"] != "" {
		params["currency_id"] = []any{"=", p["CurrencyId"]}
	}
	return params
}

func (account *Account) GetWithPaginate(db *gorm.DB, r *util.Response) {
	var results []Account
	where, values := account.WhereBuild(account.QueryParams(r.Params))
	condition := db.Model(account).Where(where, values...)
	condition.Count(&r.Pagination.Count)
	r.Pagination.Init()
	if err := condition.Order(account.TableName() + "." + r.Pagination.Order).
		Offset((int(r.Pagination.CurrentPage) - 1) * int(r.Pagination.PerPage)).Limit(int(r.Pagination.PerPage)).
		Find(&results).Error; err != nil {
		return
	}
	r.Body = results
}
