package dom

type AccountVersion struct {
	CommonModel
	ModifiableType string `gorm:"index:idx0,unique;size:16" form:"ModifiableType" validate:"oneof=salary"`
	ModifiableId   int    `gorm:"index:idx0,unique;" form:"ModifiableId"`
	AccountId      int    `gorm:"index:idx0,unique;index:idx1" form:"AccountId"`
	Fun            int    `gorm:"size:2"`
	Available      int
}

func (AccountVersion) TableName() string {
	return "account_version"
}
