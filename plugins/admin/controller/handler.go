package controller

import (
	template2 "html/template"
	"regexp"

	"github.com/wangxx2026/go-admin/context"
	"github.com/wangxx2026/go-admin/modules/auth"
	"github.com/wangxx2026/go-admin/modules/errors"
	"github.com/wangxx2026/go-admin/modules/logger"
	"github.com/wangxx2026/go-admin/plugins/admin/modules/constant"
	"github.com/wangxx2026/go-admin/plugins/admin/modules/form"
	"github.com/wangxx2026/go-admin/plugins/admin/modules/parameter"
	"github.com/wangxx2026/go-admin/plugins/admin/modules/response"
	"github.com/wangxx2026/go-admin/plugins/admin/modules/table"
	"github.com/wangxx2026/go-admin/template"
	"github.com/wangxx2026/go-admin/template/types"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// GlobalDeferHandler is a global error handler of admin plugin.
func (h *Handler) GlobalDeferHandler(ctx *context.Context) {

	logger.Access(ctx)

	if !h.config.OperationLogOff {
		h.RecordOperationLog(ctx)
	}

	if err := recover(); err != nil {
		logger.ErrorCtx(ctx, "GlobalDeferHandler error %+v", err)

		var (
			errMsg string
			ok     bool
			e      error
		)

		if errMsg, ok = err.(string); !ok {
			if e, ok = err.(error); ok {
				errMsg = e.Error()
			}
		}

		if errMsg == "" {
			errMsg = "system error"
		}

		if ctx.WantJSON() {
			response.Error(ctx, errMsg)
			return
		}

		if ok, _ = regexp.MatchString("/edit(.*)", ctx.Path()); ok {
			h.setFormWithReturnErrMessage(ctx, errMsg, "edit")
			return
		}
		if ok, _ = regexp.MatchString("/new(.*)", ctx.Path()); ok {
			h.setFormWithReturnErrMessage(ctx, errMsg, "new")
			return
		}

		h.HTML(ctx, auth.Auth(ctx), template.WarningPanelWithDescAndTitle(ctx, errMsg, errors.Msg, errors.Msg))
	}
}

func (h *Handler) setFormWithReturnErrMessage(ctx *context.Context, errMsg string, kind string) {

	var (
		formInfo table.FormInfo
		prefix   = ctx.Query(constant.PrefixKey)
		panel    = h.table(prefix, ctx)
		btnWord  template2.HTML
		f        *types.FormPanel
	)

	if kind == "edit" {
		f = panel.GetForm()
		id := ctx.Query("id")
		if id == "" {
			id = ctx.Request.MultipartForm.Value[panel.GetPrimaryKey().Name][0]
		}
		formInfo, _ = panel.GetDataWithId(parameter.GetParam(ctx.Request.URL,
			panel.GetInfo().DefaultPageSize,
			panel.GetInfo().SortField,
			panel.GetInfo().GetSort()).WithPKs(id))
		btnWord = f.FormEditBtnWord
	} else {
		f = panel.GetActualNewForm()
		formInfo = panel.GetNewFormInfo()
		formInfo.Title = f.Title
		formInfo.Description = f.Description
		btnWord = f.FormNewBtnWord
	}

	queryParam := parameter.GetParam(ctx.Request.URL, panel.GetInfo().DefaultPageSize,
		panel.GetInfo().SortField, panel.GetInfo().GetSort()).GetRouteParamStr()

	h.HTML(ctx, auth.Auth(ctx), types.Panel{
		Content: aAlert(ctx).Warning(errMsg) + formContent(ctx, aForm(ctx).
			SetContent(formInfo.FieldList).
			SetTabContents(formInfo.GroupFieldList).
			SetTabHeaders(formInfo.GroupFieldHeaders).
			SetTitle(template2.HTML(cases.Title(language.Und).String(kind))).
			SetPrimaryKey(panel.GetPrimaryKey().Name).
			SetPrefix(h.config.PrefixFixSlash()).
			SetHiddenFields(map[string]string{
				form.TokenKey:    h.authSrv().AddToken(),
				form.PreviousKey: h.config.Url("/info/" + prefix + queryParam),
			}).
			SetUrl(h.config.Url("/"+kind+"/"+prefix)).
			SetOperationFooter(formFooter(ctx, kind, f.IsHideContinueEditCheckBox, f.IsHideContinueNewCheckBox,
				f.IsHideResetButton, btnWord)).
			SetHeader(f.HeaderHtml).
			SetFooter(f.FooterHtml), len(formInfo.GroupFieldHeaders) > 0,
			ctx.IsIframe(),
			f.IsHideBackButton, f.Header),
		Description: template2.HTML(formInfo.Description),
		Title:       template2.HTML(formInfo.Title),
	})

	ctx.AddHeader(constant.PjaxUrlHeader, h.config.Url("/info/"+prefix+"/"+kind+queryParam))
}
