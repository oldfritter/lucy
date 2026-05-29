package model

import (
	"gorm.io/gorm"

	"github.com/oldfritter/lucy/dom"
	"github.com/oldfritter/lucy/util"
)

type User struct {
	dom.User

	Accounts []*Account `gorm:"foreignKey:UserId" json:"Accounts,omitempty"`
}

func (user *User) QueryParams(p map[string]string) map[string][]any {
	params := make(map[string][]any)
	if p["Nickname"] != "" {
		params["nickname"] = []any{"like", p["Nickname"]}
	}
	if p["Phone"] != "" {
		params["phone"] = []any{"=", p["Phone"]}
	}
	if p["Username"] != "" {
		params["username"] = []any{"=", p["Username"]}
	}
	return params
}

func (user *User) GetWithPaginate(db *gorm.DB, r *util.Response) {
	var results []*User
	where, values := user.WhereBuild(user.QueryParams(r.Params))
	condition := db.Model(user).Where(where, values...)
	condition.Count(&r.Pagination.Count)
	r.Pagination.Init()
	if err := condition.
		Order(user.TableName() + "." + r.Pagination.Order).
		Offset((int(r.Pagination.CurrentPage) - 1) * int(r.Pagination.PerPage)).
		Limit(int(r.Pagination.PerPage)).
		Find(&results).Error; err != nil {
		return
	}
	r.Body = results
}
