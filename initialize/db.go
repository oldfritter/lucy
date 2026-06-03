package initialize

import (
	"encoding/json"
	"log"

	"gorm.io/gorm"

	"github.com/oldfritter/lucy/dom"
	"github.com/oldfritter/lucy/internal/cache"
	"github.com/oldfritter/lucy/lib/db"
	"github.com/oldfritter/lucy/lib/kv"
	"github.com/oldfritter/lucy/model"
)

func MigrateDB() {
	AutoMigrate(db.MysqlDB)
	if err := loadApiKeysToRedis(); err != nil {
		log.Printf("[init] load api keys to redis failed: %v", err)
	}
}

func loadApiKeysToRedis() error {
	var keys []model.UserApiKey
	if err := db.MysqlDB.Where("is_active = ?", true).Preload("Product").Find(&keys).Error; err != nil {
		return err
	}

	conn := kv.GetRedisConn("data")
	defer conn.Close()

	for _, k := range keys {
		c := cache.ApiKeyCache{
			Secret:      k.Secret,
			IsActive:    k.IsActive,
			UserId:      k.UserId,
			CaptchaType: k.CaptchaType,
			PerMinuteLimit: func() int {
				if k.Product != nil {
					return k.Product.PerMinuteLimit
				}
				return 100
			}(),
		}
		data, _ := json.Marshal(c)
		conn.Do("SET", "lucy:apikey:"+k.Key, string(data))
	}
	return nil
}

func AutoMigrate(db *gorm.DB) {
	db.AutoMigrate(
		&model.Account{},
		&dom.AccountVersion{},
		&model.CaptchaImageRotate{},
		&model.CaptchaText4{},
		&model.CaptchaText5{},
		&model.CaptchaText6{},
		&model.Currency{},
		&model.User{},
		&model.UserApiKey{},
		&model.Image{},
		&model.Campaign{},
		&model.Order{},
		&model.Income{},
		&model.Refund{},
	)
}
