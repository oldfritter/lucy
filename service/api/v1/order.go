package v1

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/oldfritter/lucy/base"
	"github.com/oldfritter/lucy/dom"
	"github.com/oldfritter/lucy/lib/db"
	"github.com/oldfritter/lucy/model"
	"github.com/oldfritter/lucy/util"
)

// CreateOrder 创建订单并同时创建关联的 Reason
func CreateOrder(c echo.Context) (err error) {
	claims, _ := base.GetClaim(c)

	var input struct {
		CurrencyId     int    `form:"CurrencyId" validate:"required"`
		Amount         int    `form:"Amount" validate:"required,min=1"`
		DeductedAmount int    `form:"DeductedAmount"`
		ReasonType     string `form:"ReasonType" validate:"required"`
		Name              string `form:"Name"`
		CaptchaType       string `form:"CaptchaType"`
		BackgroundImages  string `form:"BackgroundImages"`
		WordBank          string `form:"WordBank"`
		UseSystemWordBank bool   `form:"UseSystemWordBank"`
		CaptchaCount      int    `form:"CaptchaCount"`
	}
	if err = c.Bind(&input); err != nil {
		return util.BuildError("1001")
	}
	if err = c.Validate(&input); err != nil {
		return util.BuildError("1002", err.Error())
	}

	if input.ReasonType != "Campaign" {
		return util.BuildError("1005", "不支持的 ReasonType")
	}
	if input.Name == "" {
		return util.BuildError("1002", "Campaign Name 不能为空")
	}
	if input.CaptchaType == "" {
		input.CaptchaType = "text4"
	}
	if input.CaptchaCount < 1 {
		input.CaptchaCount = 1
	}

	finalAmount := input.Amount - input.DeductedAmount
	if finalAmount < 0 {
		finalAmount = 0
	}
	orderNo := fmt.Sprintf("LC%s%s", time.Now().Format("20060102150405"), util.RandNumberStringRunes(6))

	tx := db.BeginTx()
	defer tx.DbRollback()

	campaign := model.Campaign{
		Campaign: dom.Campaign{
			UserId:            claims.UserId,
			Name:              input.Name,
			CaptchaType:       input.CaptchaType,
			BackgroundImages:  input.BackgroundImages,
			WordBank:          input.WordBank,
			UseSystemWordBank: input.UseSystemWordBank,
			CaptchaCount:      input.CaptchaCount,
			Status:            dom.StatusPending,
		},
	}
	if err = tx.Create(&campaign).Error; err != nil {
		return util.BuildError("1007", "创建 Campaign 失败")
	}

	order := model.Order{
		Order: dom.Order{
			UserId:         claims.UserId,
			CurrencyId:     input.CurrencyId,
			Amount:         input.Amount,
			DeductedAmount: input.DeductedAmount,
			FinalAmount:    finalAmount,
			OrderNo:        orderNo,
			Status:         dom.StatusPending,
			ReasonType:     input.ReasonType,
			ReasonId:       campaign.Id,
		},
	}
	if err = tx.Create(&order).Error; err != nil {
		return util.BuildError("1007", "创建订单失败")
	}

	tx.DbCommit()

	db.MysqlDB.Preload("Currency").First(&order, order.Id)
	db.MysqlDB.First(&campaign, campaign.Id)

	response := util.SuccessResponse()
	response.Body = map[string]any{
		"order":    order,
		"campaign": campaign,
	}
	return c.JSON(http.StatusOK, response)
}

func GetMyOrderList(c echo.Context) (err error) {
	claims, _ := base.GetClaim(c)
	var order model.Order
	body := util.ArrayResponse()
	if err = c.Bind(&body.Params); err != nil {
		return util.BuildError("1001")
	}
	if err = c.Bind(body.Pagination); err != nil {
		return util.BuildError("1001")
	}
	body.Params["UserId"] = strconv.Itoa(claims.UserId)
	order.GetWithPaginate(db.MysqlDB, &body)
	return c.JSON(http.StatusOK, &body)
}

func GetMyOrder(c echo.Context) (err error) {
	claims, _ := base.GetClaim(c)
	var order model.Order
	if err = db.MysqlDB.Preload("Currency").Preload("Incomes").
		Where("id = ? AND user_id = ?", c.Param("id"), claims.UserId).
		First(&order).Error; err != nil {
		return util.BuildError("1003", "订单不存在")
	}
	response := util.SuccessResponse()
	response.Body = order
	return c.JSON(http.StatusOK, response)
}
