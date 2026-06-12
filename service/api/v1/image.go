package v1

import (
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/oldfritter/lucy/base"
	"github.com/oldfritter/lucy/dom"
	"github.com/oldfritter/lucy/lib/db"
	"github.com/oldfritter/lucy/lib/storage/oss"
	"github.com/oldfritter/lucy/model"
	"github.com/oldfritter/lucy/util"
)

var allowedMimeTypes = map[string]bool{
	"image/jpeg":    true,
	"image/png":     true,
	"image/gif":     true,
	"image/webp":    true,
	"image/bmp":     true,
	"image/svg+xml": true,
	"text/xml":      true, // SVG 可能被识别为 XML
}

// UploadImage 用户上传图片
func UploadImage(c echo.Context) (err error) {
	claims, _ := base.GetClaim(c)

	// 读取上传文件
	file, header, err := c.Request().FormFile("file")
	if err != nil {
		return util.BuildError("1001")
	}
	defer file.Close()

	// 读取文件字节
	data, err := io.ReadAll(file)
	if err != nil {
		return util.BuildError("1001")
	}
	if len(data) == 0 {
		return util.BuildError("1001")
	}

	// 检测 MIME 类型
	mimeType := http.DetectContentType(data)
	if !allowedMimeTypes[mimeType] {
		return util.BuildError("1504")
	}

	// 扩展名
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext == "" {
		ext = mimeToExt(mimeType)
	}

	// 从表单获取可选的描述
	desc := c.FormValue("Desc")

	img := model.Image{
		Image: dom.Image{
			UserId:    claims.UserId,
			Name:      header.Filename,
			Size:      header.Size,
			MimeType:  mimeType,
			Extension: ext,
			Desc:      desc,
		},
	}

	tx := db.BeginTx()
	defer tx.DbRollback()

	if err = tx.Create(&img).Error; err != nil {
		return util.BuildError("1005")
	}

	// DB 写入成功后立即上传至 OSS
	if _, err = oss.PutObject(img.Path, &data); err != nil {
		return util.BuildError("1505")
	}

	tx.DbCommit()

	img.Url, _ = oss.GetObjectURL(img.Path, 3600)
	response := util.SuccessResponse()
	response.Body = img
	return c.JSON(http.StatusOK, response)
}

// GetMyImageList 当前用户查看自己的图片列表
func GetMyImageList(c echo.Context) (err error) {
	claims, _ := base.GetClaim(c)
	var (
		img  model.Image
		body = util.ArrayResponse()
	)
	if err = c.Bind(&body.Params); err != nil {
		return util.BuildError("1001")
	}
	if err = c.Bind(body.Pagination); err != nil {
		return util.BuildError("1001")
	}
	body.Params["UserId"] = ""
	db.MysqlDB.Model(&img).
		Where("user_id = ?", claims.UserId).
		Count(&body.Pagination.Count)
	body.Pagination.Init()

	var results []model.Image
	if err = db.MysqlDB.Where("user_id = ?", claims.UserId).
		Order("id DESC").
		Offset((int(body.Pagination.CurrentPage) - 1) * int(body.Pagination.PerPage)).
		Limit(int(body.Pagination.PerPage)).
		Find(&results).Error; err != nil {
		return util.BuildError("1000")
	}
	for i := range results {
		results[i].Url, _ = oss.GetObjectURL(results[i].Path, 3600)
	}
	body.Body = results
	return c.JSON(http.StatusOK, &body)
}

// GetMyImage 当前用户查看自己的单张图片详情
func GetMyImage(c echo.Context) (err error) {
	claims, _ := base.GetClaim(c)
	var img model.Image
	if err = db.MysqlDB.Where("id = ? AND user_id = ?", c.Param("id"), claims.UserId).
		First(&img).Error; err != nil {
		return util.BuildError("1503")
	}
	img.Url, _ = oss.GetObjectURL(img.Path, 3600)
	response := util.SuccessResponse()
	response.Body = img
	return c.JSON(http.StatusOK, response)
}

// DeleteMyImage 当前用户删除自己的图片（同时删除 OSS 文件）
func DeleteMyImage(c echo.Context) (err error) {
	claims, _ := base.GetClaim(c)
	var img model.Image
	if err = db.MysqlDB.Where("id = ? AND user_id = ?", c.Param("id"), claims.UserId).
		First(&img).Error; err != nil {
		return util.BuildError("1503")
	}

	// 先删 OSS，再删 DB；OSS 失败则保留 DB 记录
	if err = oss.DeleteObject(img.Path); err != nil {
		return util.BuildError("1505")
	}

	tx := db.BeginTx()
	defer tx.DbRollback()
	tx.Delete(&img)
	tx.DbCommit()

	response := util.SuccessResponse()
	return c.JSON(http.StatusOK, response)
}

// mimeToExt MIME 类型到默认扩展名映射
func mimeToExt(mime string) string {
	switch mime {
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/gif":
		return ".gif"
	case "image/webp":
		return ".webp"
	case "image/bmp":
		return ".bmp"
	case "image/svg+xml":
		return ".svg"
	default:
		return ".bin"
	}
}
