package model

import (
	"gorm.io/gorm"

	"github.com/oldfritter/lucy/dom"
	"github.com/oldfritter/lucy/util"
)

type Campaign struct {
	dom.Campaign

	User *User `gorm:"foreignKey:UserId" json:",omitempty"`
}

func (c *Campaign) QueryParams(p map[string]string) map[string][]any {
	params := make(map[string][]any)
	if p["UserId"] != "" {
		params["user_id"] = []any{"=", p["UserId"]}
	}
	if p["Name"] != "" {
		params["name"] = []any{"like", p["Name"]}
	}
	if p["Status"] != "" {
		params["status"] = []any{"=", p["Status"]}
	}
	if p["CaptchaType"] != "" {
		params["captcha_type"] = []any{"=", p["CaptchaType"]}
	}
	return params
}

func (c *Campaign) GetWithPaginate(db *gorm.DB, r *util.Response) {
	var results []Campaign
	where, values := c.WhereBuild(c.QueryParams(r.Params))
	condition := db.Model(c).Where(where, values...)
	condition.Count(&r.Pagination.Count)
	r.Pagination.Init()
	if err := condition.
		Order(c.TableName() + "." + r.Pagination.Order).
		Offset((int(r.Pagination.CurrentPage) - 1) * int(r.Pagination.PerPage)).
		Limit(int(r.Pagination.PerPage)).
		Find(&results).Error; err != nil {
		return
	}
	r.Body = results
}
