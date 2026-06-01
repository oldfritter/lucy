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

// CreateOrder 依据 CampaignId 查出关联商品定价，据此生成订单
func CreateOrder(c echo.Context) (err error) {
	claims, _ := base.GetClaim(c)

	var input struct {
		CampaignId int `form:"CampaignId" validate:"required"`
	}
	if err = c.Bind(&input); err != nil {
		return util.BuildError("1001")
	}
	if err = c.Validate(&input); err != nil {
		return util.BuildError("1002", err.Error())
	}

	// 验证 Campaign 存在且属于当前用户，同时加载关联商品
	var campaign model.Campaign
	if err = db.MysqlDB.Preload("Product").Where("id = ? AND user_id = ?", input.CampaignId, claims.UserId).
		First(&campaign).Error; err != nil {
		return util.BuildError("1003", "投放不存在")
	}
	if campaign.Status != dom.StatusPending {
		return util.BuildError("1002", "投放状态不允许创建订单")
	}
	if campaign.Product == nil {
		return util.BuildError("1003", "投放关联商品不存在")
	}

	amount := campaign.Product.Amount
	finalAmount := amount
	orderNo := fmt.Sprintf("LC%s%s", time.Now().Format("20060102150405"), util.RandNumberStringRunes(6))

	tx := db.BeginTx()
	defer tx.DbRollback()

	order := model.Order{
		Order: dom.Order{
			UserId:         claims.UserId,
			CurrencyId:     campaign.Product.CurrencyId,
			Amount:         amount,
			DeductedAmount: 0,
			FinalAmount:    finalAmount,
			OrderNo:        orderNo,
			Status:         dom.StatusPending,
			ReasonType:     "Campaign",
			ReasonId:       campaign.Id,
		},
	}
	if err = tx.Create(&order).Error; err != nil {
		return util.BuildError("1007", "创建订单失败")
	}

	tx.DbCommit()

	db.MysqlDB.Preload("Currency").First(&order, order.Id)

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
	if err = db.MysqlDB.Preload("Currency").Preload("Incomes").Preload("Refunds").
		Where("id = ? AND user_id = ?", c.Param("id"), claims.UserId).
		First(&order).Error; err != nil {
		return util.BuildError("1003", "订单不存在")
	}
	response := util.SuccessResponse()
	response.Body = order
	return c.JSON(http.StatusOK, response)
}
