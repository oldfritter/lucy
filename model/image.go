package model

import (
	"gorm.io/gorm"

	"github.com/oldfritter/lucy/dom"
	"github.com/oldfritter/lucy/util"
)

type Image struct {
	dom.Image

	User *User `gorm:"foreignKey:UserId" json:",omitempty"`
}

func (img *Image) GetWithPaginate(db *gorm.DB, r *util.Response) {
	var results []*Image
	where, values := img.WhereBuild(img.QueryParams(r.Params))
	condition := db.Model(img).Where(where, values...)
	condition.Count(&r.Pagination.Count)
	r.Pagination.Init()
	if err := condition.
		Order(img.TableName() + "." + r.Pagination.Order).
		Offset((int(r.Pagination.CurrentPage) - 1) * int(r.Pagination.PerPage)).
		Limit(int(r.Pagination.PerPage)).
		Find(&results).Error; err != nil {
		return
	}
	r.Body = results
}
