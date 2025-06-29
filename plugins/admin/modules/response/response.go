package response

import (
	"net/http"

	"github.com/wangxx2026/go-admin/context"
	"github.com/wangxx2026/go-admin/modules/auth"
	"github.com/wangxx2026/go-admin/modules/config"
	"github.com/wangxx2026/go-admin/modules/db"
	"github.com/wangxx2026/go-admin/modules/errors"
	"github.com/wangxx2026/go-admin/modules/language"
	"github.com/wangxx2026/go-admin/modules/menu"
	"github.com/wangxx2026/go-admin/template"
	"github.com/wangxx2026/go-admin/template/types"
)

func Ok(ctx *context.Context) {
	ctx.JSON(http.StatusOK, map[string]interface{}{
		"code": http.StatusOK,
		"msg":  "ok",
	})
}

func OkWithMsg(ctx *context.Context, msg string) {
	ctx.JSON(http.StatusOK, map[string]interface{}{
		"code": http.StatusOK,
		"msg":  msg,
	})
}

func OkWithData(ctx *context.Context, data map[string]interface{}) {
	ctx.JSON(http.StatusOK, map[string]interface{}{
		"code": http.StatusOK,
		"msg":  "ok",
		"data": data,
	})
}

func BadRequest(ctx *context.Context, msg string) {
	ctx.JSON(http.StatusBadRequest, map[string]interface{}{
		"code": http.StatusBadRequest,
		"msg":  language.Get(msg),
	})
}

func Alert(ctx *context.Context, desc, title, msg string, conn db.Connection, btns *types.Buttons,
	pageType ...template.PageType) {
	user := auth.Auth(ctx)

	pt := template.Error500Page
	if len(pageType) > 0 {
		pt = pageType[0]
	}

	pageTitle, description, content := template.GetPageContentFromPageType(ctx, title, desc, msg, pt)

	tmpl, tmplName := template.Default(ctx).GetTemplate(ctx.IsPjax())
	buf := template.Execute(ctx, &template.ExecuteParam{
		User:     user,
		TmplName: tmplName,
		Tmpl:     tmpl,
		Panel: types.Panel{
			Content:     content,
			Description: description,
			Title:       pageTitle,
		},
		Config:    config.Get(),
		Menu:      menu.GetGlobalMenu(user, conn, ctx.Lang()).SetActiveClass(config.URLRemovePrefix(ctx.Path())),
		Animation: true,
		Buttons:   *btns,
		IsPjax:    ctx.IsPjax(),
		Iframe:    ctx.IsIframe(),
	})
	ctx.HTML(http.StatusOK, buf.String())
}

func Error(ctx *context.Context, msg string, datas ...map[string]interface{}) {
	res := map[string]interface{}{
		"code": http.StatusInternalServerError,
		"msg":  language.Get(msg),
	}
	if len(datas) > 0 {
		res["data"] = datas[0]
	}
	ctx.JSON(http.StatusInternalServerError, res)
}

func Denied(ctx *context.Context, msg string) {
	ctx.JSON(http.StatusInternalServerError, map[string]interface{}{
		"code": http.StatusForbidden,
		"msg":  language.Get(msg),
	})
}

var OffLineHandler = func(ctx *context.Context) {
	if config.GetSiteOff() {
		if ctx.WantHTML() {
			ctx.HTML(http.StatusOK, `<html><body><h1>The website is offline</h1></body></html>`)
		} else {
			ctx.JSON(http.StatusForbidden, map[string]interface{}{
				"code": http.StatusForbidden,
				"msg":  language.Get(errors.SiteOff),
			})
		}
		ctx.Abort()
	}
}
