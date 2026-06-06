package model

import (
	"gorm.io/gorm"

	"github.com/oldfritter/lucy/dom"
	"github.com/oldfritter/lucy/internal/cache"
	"github.com/oldfritter/lucy/util"
)

type UserApiKey struct {
	dom.UserApiKey

	User    *User    `gorm:"foreignKey:UserId" json:",omitempty"`
	Product *Product `gorm:"foreignKey:ProductId" json:",omitempty"`
}

func (uak *UserApiKey) AfterCreate(tx *gorm.DB) error {
	return syncApiKeyCache(uak)
}

func (uak *UserApiKey) AfterUpdate(tx *gorm.DB) error {
	return syncApiKeyCache(uak)
}

func (uak *UserApiKey) AfterDelete(tx *gorm.DB) error {
	_ = cache.DelApiKeyCache(uak.Key)
	return nil
}

func syncApiKeyCache(uak *UserApiKey) error {
	return cache.SetApiKeyCache(uak.Key, &cache.ApiKeyCache{
		Id:          uak.Id,
		Secret:      uak.Secret,
		IsActive:    uak.IsActive,
		UserId:      uak.UserId,
		CaptchaType: uak.CaptchaType,
		PerMinuteLimit: func() int {
			if uak.Product != nil {
				return uak.Product.PerMinuteLimit
			}
			return 100
		}(),
	})
}

// GenerateKeys 使用密码学安全随机数生成 32 位 Key 与 Secret
func (uak *UserApiKey) GenerateKeys() {
	uak.Key = util.RandStringRunes(32)
	uak.Secret = util.RandStringRunes(32)
}

// Mask 对 Key 和 Secret 做部分脱敏，用于列表等场景
func (uak *UserApiKey) Mask() {
	if len(uak.Key) > 8 {
		uak.Key = uak.Key[:4] + "****" + uak.Key[len(uak.Key)-4:]
	}
	if len(uak.Secret) > 8 {
		uak.Secret = uak.Secret[:4] + "****" + uak.Secret[len(uak.Secret)-4:]
	} else if len(uak.Secret) > 0 {
		uak.Secret = "****"
	}
}

func (uak *UserApiKey) QueryParams(p map[string]string) map[string][]any {
	params := make(map[string][]any)
	if p["UserId"] != "" {
		params["user_id"] = []any{"=", p["UserId"]}
	}

	if p["CaptchaType"] != "" {
		params["captcha_type"] = []any{"=", p["CaptchaType"]}
	}
	if p["Name"] != "" {
		params["name"] = []any{"like", p["Name"]}
	}
	if p["IsActive"] != "" {
		params["is_active"] = []any{"=", p["IsActive"]}
	}
	return params
}

func (uak *UserApiKey) GetWithPaginate(db *gorm.DB, r *util.Response) {
	var results []UserApiKey
	where, values := uak.WhereBuild(uak.QueryParams(r.Params))
	condition := db.Model(uak).Where(where, values...)
	condition.Count(&r.Pagination.Count)
	r.Pagination.Init()
	if err := condition.
		Order(uak.TableName() + "." + r.Pagination.Order).
		Offset((int(r.Pagination.CurrentPage) - 1) * int(r.Pagination.PerPage)).
		Limit(int(r.Pagination.PerPage)).
		Find(&results).Error; err != nil {
		return
	}
	// 列表场景脱敏
	for i := range results {
		results[i].Mask()
	}
	r.Body = results
}
