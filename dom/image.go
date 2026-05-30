package dom

import (
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/oldfritter/lucy/util"
)

type Image struct {
	CommonModel
	UserId    int    `gorm:"index" form:"UserId" query:"UserId"`                     // 所属用户
	Name      string `gorm:"size:255" form:"Name"`                                   // 原始文件名
	Path      string `gorm:"size:512" query:"Path"`                                  // OSS 存储路径
	Size      int64  `gorm:"default:0" form:"Size" query:"Size"`                     // 文件大小（字节）
	MimeType  string `gorm:"size:64" form:"MimeType" query:"MimeType"`               // MIME 类型，如 image/jpeg
	Extension string `gorm:"size:16" form:"Extension" query:"Extension"`             // 文件扩展名，如 .jpg
	Width     int    `gorm:"default:0" form:"Width" query:"Width"`                   // 图片宽度（像素）
	Height    int    `gorm:"default:0" form:"Height" query:"Height"`                 // 图片高度（像素）
	Desc      string `gorm:"size:1024" form:"Desc" query:"Desc"`                     // 描述信息
	Url       string `gorm:"-" json:"Url,omitempty"`                                 // 带签名的临时下载地址（不持久化）
}

func (*Image) TableName() string {
	return "image"
}

func (img *Image) BeforeCreate(db *gorm.DB) (err error) {
	if img.Path == "" {
		img.Path = fmt.Sprintf("upload/%s/%s%s", time.Now().Format("2006/01/02"), util.RandStringRunes(32), img.Extension)
	}
	return
}

func (img *Image) QueryParams(p map[string]string) map[string][]any {
	params := make(map[string][]any)
	if p["UserId"] != "" {
		params["user_id"] = []any{"=", p["UserId"]}
	}
	if p["Name"] != "" {
		params["name"] = []any{"like", p["Name"]}
	}
	if p["Path"] != "" {
		params["path"] = []any{"like", p["Path"]}
	}
	if p["MimeType"] != "" {
		params["mime_type"] = []any{"=", p["MimeType"]}
	}
	if p["Extension"] != "" {
		params["extension"] = []any{"=", p["Extension"]}
	}
	return params
}
