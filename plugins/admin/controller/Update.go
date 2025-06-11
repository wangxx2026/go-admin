package controller

import (
	"github.com/wangxx2026/go-admin/context"
	"github.com/wangxx2026/go-admin/plugins/admin/modules/guard"
	"github.com/wangxx2026/go-admin/plugins/admin/modules/response"
)

// Update update the table row of given id.
func (h *Handler) Update(ctx *context.Context) {

	param := guard.GetUpdateParam(ctx)

	err := param.Panel.UpdateData(ctx, param.Value)

	if err != nil {
		response.Error(ctx, err.Error())
		return
	}

	response.Ok(ctx)
}
