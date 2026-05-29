package initialize

import (
	"gorm.io/gorm"

	"github.com/oldfritter/lucy/dom"
	"github.com/oldfritter/lucy/lib/db"
	"github.com/oldfritter/lucy/model"
)

func MigrateDB() {
	AutoMigrate(db.MysqlDB)
}

func AutoMigrate(db *gorm.DB) {
	db.AutoMigrate(
		&model.Account{},
		&dom.AccountVersion{},
		&model.CaptchaText4Image{},
		&model.CaptchaText5Image{},
		&model.CaptchaText6Image{},
		&model.Currency{},
		&model.User{},
	)
}
