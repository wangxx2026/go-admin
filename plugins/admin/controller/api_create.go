package controller

import (
	"github.com/wangxx2026/go-admin/context"
	"github.com/wangxx2026/go-admin/modules/file"
	"github.com/wangxx2026/go-admin/plugins/admin/modules/constant"
	"github.com/wangxx2026/go-admin/plugins/admin/modules/guard"
	"github.com/wangxx2026/go-admin/plugins/admin/modules/response"
)

func (h *Handler) ApiCreate(ctx *context.Context) {
	param := guard.GetNewFormParam(ctx)

	if len(param.MultiForm.File) > 0 {
		err := file.GetFileEngine(h.config.FileUploadEngine.Name).Upload(param.MultiForm)
		if err != nil {
			response.Error(ctx, err.Error())
			return
		}
	}

	err := param.Panel.InsertData(ctx, param.Value())
	if err != nil {
		response.Error(ctx, err.Error())
		return
	}

	response.Ok(ctx)
}

func (h *Handler) ApiCreateForm(ctx *context.Context) {

	var (
		params           = guard.GetShowNewFormParam(ctx)
		prefix, paramStr = params.Prefix, params.Param.GetRouteParamStr()
		panel            = h.table(prefix, ctx)
		formInfo         = panel.GetNewFormInfo()
		infoUrl          = h.routePathWithPrefix("api_info", prefix) + paramStr
		newUrl           = h.routePathWithPrefix("api_new", prefix)
		referer          = ctx.Referer()
		f                = panel.GetActualNewForm()
	)

	if referer != "" && !isInfoUrl(referer) && !isNewUrl(referer, ctx.Query(constant.PrefixKey)) {
		infoUrl = referer
	}

	response.OkWithData(ctx, map[string]interface{}{
		"panel": formInfo,
		"urls": map[string]string{
			"info": infoUrl,
			"new":  newUrl,
		},
		"pk":     panel.GetPrimaryKey().Name,
		"header": f.HeaderHtml,
		"footer": f.FooterHtml,
		"prefix": h.config.PrefixFixSlash(),
		"token":  h.authSrv().AddToken(),
		"operation_footer": formFooter(ctx, "new", f.IsHideContinueEditCheckBox, f.IsHideContinueNewCheckBox,
			f.IsHideResetButton, f.FormNewBtnWord),
	})
}
