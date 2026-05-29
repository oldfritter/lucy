package helper

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"

	lang "github.com/oldfritter/lucy/lib/language"
	"github.com/oldfritter/lucy/util"
)

func CustomHTTPErrorHandler(err error, context echo.Context) {
	if err.Error() == "Code=404, message=Not Found" {
		context.JSON(http.StatusOK, util.BuildError("1000", "亮出你的獠牙，拱开这片土地，你就有食吃了 ！"))
		return
	}
	language := "zh-CN"
	if context.Get("Language") != nil {
		if t, ok := context.Get("Language").(string); ok {
			language = t
		}
	}
	if response, ok := err.(util.Response); ok {
		if response.Head["Code"] != "" && response.Head["Msg"] == "" {
			response.Head["Msg"] = fmt.Sprint(lang.I18n.T(language, "Code."+response.Head["Code"]))
		}
		context.JSON(http.StatusBadRequest, response)
	} else {
		context.JSON(http.StatusInternalServerError, response)
	}
	return
}
