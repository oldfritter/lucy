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
		return verifyTextimage4(c, req)
	case 5:
		return verifyTextimage5(c, req)
	case 6:
		return verifyTextimage6(c, req)
	}
	return util.BuildError("1001")
}

func verifyTextimage4(c echo.Context, req verifyRequest) error {
	var captcha model.CaptchaTextimage4
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

func verifyTextimage5(c echo.Context, req verifyRequest) error {
	var captcha model.CaptchaTextimage5
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

func verifyTextimage6(c echo.Context, req verifyRequest) error {
	var captcha model.CaptchaTextimage6
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
