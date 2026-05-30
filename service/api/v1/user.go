package v1

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/oldfritter/lucy/base"
	"github.com/oldfritter/lucy/dom"
	"github.com/oldfritter/lucy/lib/db"
	"github.com/oldfritter/lucy/lib/kv"
	"github.com/oldfritter/lucy/model"
	"github.com/oldfritter/lucy/util"
)

type registerRequest struct {
	Username string `form:"Username" validate:"required,max=32"`
	Password string `form:"Password" validate:"required,max=32"`
	Nickname string `form:"Nickname" validate:"max=32"`
	Email    string `form:"Email" validate:"email"`
	Phone    string `form:"Phone"`
	Gender   string `form:"Gender" validate:"oneof=f m"`
}

type loginRequest struct {
	Username string `form:"Username" validate:"required"`
	Password string `form:"Password" validate:"required"`
}

type userResponse struct {
	Id       int    `json:"Id"`
	Username string `json:"Username"`
	Nickname string `json:"Nickname"`
	Phone    string `json:"Phone"`
	Email    string `json:"Email"`
	Gender   string `json:"Gender"`
	Locale   string `json:"Locale"`
	Token    string `json:"Token,omitempty"`
}

// sanitizeUser 清除敏感字段并构造响应体
func sanitizeUser(user *model.User) userResponse {
	return userResponse{
		Id:       user.Id,
		Username: user.Username,
		Nickname: user.Nickname,
		Phone:    user.Phone,
		Email:    user.Email,
		Gender:   user.Gender,
		Locale:   user.Locale,
	}
}

// Register 用户注册
func Register(c echo.Context) (err error) {
	var req registerRequest
	if err = c.Bind(&req); err != nil {
		return util.BuildError("1001")
	}

	// Gender 统一转小写（必须在 Validate 之前）
	if req.Gender != "" {
		req.Gender = strings.ToLower(req.Gender)
	}

	if err = c.Validate(&req); err != nil {
		return util.BuildError("1002", err.Error())
	}

	// 检查用户名是否已存在
	var exist model.User
	if db.MysqlDB.Where("username = ?", req.Username).First(&exist).Error == nil {
		return util.BuildError("1007", "用户名已存在")
	}

	user := model.User{
		User: dom.User{
			Username: req.Username,
			Password: req.Password,
			Nickname: req.Nickname,
			Email:    req.Email,
			Phone:    req.Phone,
			Gender:   req.Gender,
		},
	}

	tx := db.BeginTx()
	defer tx.DbRollback()
	if err = tx.Create(&user).Error; err != nil {
		return util.BuildError("1007")
	}
	tx.DbCommit()

	// 生成 JWT 并写入 Redis
	token, err := base.JWT.GenerateToken(&user)
	if err != nil {
		return util.BuildError("1007")
	}
	cacheRedis := kv.GetRedisConn("cache")
	defer cacheRedis.Close()
	key := "lucy:login:token:web:" + strconv.Itoa(user.Id)
	cacheRedis.Do("SET", key, token)

	resp := sanitizeUser(&user)
	resp.Token = token

	response := util.SuccessResponse()
	response.Body = resp
	return c.JSON(http.StatusOK, response)
}

// Login 用户登录
func Login(c echo.Context) (err error) {
	var req loginRequest
	if err = c.Bind(&req); err != nil {
		return util.BuildError("1001")
	}
	if err = c.Validate(&req); err != nil {
		return util.BuildError("1002")
	}

	var user model.User
	if err = db.MysqlDB.Where("username = ?", req.Username).First(&user).Error; err != nil {
		return util.BuildError("1005", "用户名或密码错误")
	}

	// 设置明文密码用于比对
	user.Password = req.Password
	if !user.CompareHashAndPassword() {
		return util.BuildError("1005", "用户名或密码错误")
	}

	// 生成 JWT 并写入 Redis
	token, err := base.JWT.GenerateToken(&user)
	if err != nil {
		return util.BuildError("1007")
	}
	cacheRedis := kv.GetRedisConn("cache")
	defer cacheRedis.Close()
	key := "lucy:login:token:web:" + strconv.Itoa(user.Id)
	cacheRedis.Do("SET", key, token)

	resp := sanitizeUser(&user)
	resp.Token = token

	response := util.SuccessResponse()
	response.Body = resp
	return c.JSON(http.StatusOK, response)
}

// GetMyProfile 查看当前用户个人信息
func GetMyProfile(c echo.Context) (err error) {
	claims, _ := base.GetClaim(c)
	var user model.User
	if err = db.MysqlDB.Where("id = ?", claims.UserId).First(&user).Error; err != nil {
		return util.BuildError("1003")
	}
	response := util.SuccessResponse()
	response.Body = sanitizeUser(&user)
	return c.JSON(http.StatusOK, response)
}

// UpdateMyProfile 修改当前用户个人信息
func UpdateMyProfile(c echo.Context) (err error) {
	claims, _ := base.GetClaim(c)
	var user model.User
	if err = db.MysqlDB.Where("id = ?", claims.UserId).First(&user).Error; err != nil {
		return util.BuildError("1003")
	}

	var input struct {
		Nickname string `form:"Nickname" validate:"max=32"`
		Email    string `form:"Email" validate:"email"`
		Phone    string `form:"Phone"`
		Gender   string `form:"Gender" validate:"oneof=f m"`
	}
	if err = c.Bind(&input); err != nil {
		return util.BuildError("1001")
	}

	updates := map[string]any{}
	if input.Nickname != "" {
		updates["nickname"] = input.Nickname
	}
	if input.Email != "" {
		updates["email"] = input.Email
	}
	if input.Phone != "" {
		updates["phone"] = input.Phone
	}
	if input.Gender != "" {
		updates["gender"] = input.Gender
	}
	if len(updates) > 0 {
		db.MysqlDB.Model(&user).Updates(updates)
		db.MysqlDB.First(&user, user.Id)
	}

	response := util.SuccessResponse()
	response.Body = sanitizeUser(&user)
	return c.JSON(http.StatusOK, response)
}
