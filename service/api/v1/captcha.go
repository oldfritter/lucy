package v1

import (
	"math"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/oldfritter/lucy/lib/db"
	"github.com/oldfritter/lucy/model"
	"github.com/oldfritter/lucy/util"
)

const verifyTolerance = 30 // 坐标验证容差（像素）

type verifyPoint struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type verifyRequest struct {
	Captcha string        `json:"captcha"`
	Points  []verifyPoint `json:"points"`
}

func VerifyCaptcha(c echo.Context) (err error) {
	var req verifyRequest
	if err = c.Bind(&req); err != nil {
		return util.BuildError("1001")
	}

	if len(req.Points) < 4 || len(req.Points) > 6 {
		return util.BuildError("1001")
	}

	switch len(req.Points) {
	case 4:
		return verifyText4Image(c, req)
	case 5:
		return verifyText5Image(c, req)
	case 6:
		return verifyText6Image(c, req)
	}
	return util.BuildError("1001")
}

func verifyText4Image(c echo.Context, req verifyRequest) error {
	var captcha model.CaptchaText4
	if err := db.MysqlDB.Where("captcha = ?", req.Captcha).First(&captcha).Error; err != nil {
		return util.BuildError("1003")
	}
	expected := [][]int{
		{captcha.Verify1X, captcha.Verify1Y},
		{captcha.Verify2X, captcha.Verify2Y},
		{captcha.Verify3X, captcha.Verify3Y},
		{captcha.Verify4X, captcha.Verify4Y},
	}
	if !matchPoints(req.Points, expected) {
		return util.BuildError("1008")
	}
	return c.JSON(http.StatusOK, util.SuccessResponse())
}

func verifyText5Image(c echo.Context, req verifyRequest) error {
	var captcha model.CaptchaText5
	if err := db.MysqlDB.Where("captcha = ?", req.Captcha).First(&captcha).Error; err != nil {
		return util.BuildError("1003")
	}
	expected := [][]int{
		{captcha.Verify1X, captcha.Verify1Y},
		{captcha.Verify2X, captcha.Verify2Y},
		{captcha.Verify3X, captcha.Verify3Y},
		{captcha.Verify4X, captcha.Verify4Y},
		{captcha.Verify5X, captcha.Verify5Y},
	}
	if !matchPoints(req.Points, expected) {
		return util.BuildError("1008")
	}
	return c.JSON(http.StatusOK, util.SuccessResponse())
}

func verifyText6Image(c echo.Context, req verifyRequest) error {
	var captcha model.CaptchaText6
	if err := db.MysqlDB.Where("captcha = ?", req.Captcha).First(&captcha).Error; err != nil {
		return util.BuildError("1003")
	}
	expected := [][]int{
		{captcha.Verify1X, captcha.Verify1Y},
		{captcha.Verify2X, captcha.Verify2Y},
		{captcha.Verify3X, captcha.Verify3Y},
		{captcha.Verify4X, captcha.Verify4Y},
		{captcha.Verify5X, captcha.Verify5Y},
		{captcha.Verify6X, captcha.Verify6Y},
	}
	if !matchPoints(req.Points, expected) {
		return util.BuildError("1008")
	}
	return c.JSON(http.StatusOK, util.SuccessResponse())
}

// rotateVerifyRequest 旋转验证码验证请求
type rotateVerifyRequest struct {
	Captcha string  `json:"captcha"`
	Angle   float64 `json:"angle"`
}

// VerifyRotateCaptcha 验证旋转验证码
func VerifyRotateCaptcha(c echo.Context) (err error) {
	var req rotateVerifyRequest
	if err = c.Bind(&req); err != nil {
		return util.BuildError("1001")
	}
	if req.Captcha == "" {
		return util.BuildError("1001")
	}

	var captcha model.CaptchaRotateImage
	if err = db.MysqlDB.Where("captcha = ?", req.Captcha).First(&captcha).Error; err != nil {
		return util.BuildError("1003")
	}
	if !captcha.Verify(map[string]any{"angle": req.Angle}) {
		return util.BuildError("1008")
	}
	return c.JSON(http.StatusOK, util.SuccessResponse())
}

func matchPoints(input []verifyPoint, expected [][]int) bool {
	for i, p := range input {
		dx := float64(p.X - expected[i][0])
		dy := float64(p.Y - expected[i][1])
		if math.Sqrt(dx*dx+dy*dy) > verifyTolerance {
			return false
		}
	}
	return true
}
