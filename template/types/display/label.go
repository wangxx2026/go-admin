package display

import (
	"github.com/wangxx2026/go-admin/context"
	"github.com/wangxx2026/go-admin/template"
	"github.com/wangxx2026/go-admin/template/types"
)

type Label struct {
	types.BaseDisplayFnGenerator
}

func init() {
	types.RegisterDisplayFnGenerator("label", new(Label))
}

func (label *Label) Get(ctx *context.Context, args ...interface{}) types.FieldFilterFn {
	return func(value types.FieldModel) interface{} {
		params := args[0].([]types.FieldLabelParam)
		if len(params) == 0 {
			return template.Default(ctx).Label().
				SetContent(template.HTML(value.Value)).
				SetType("success").
				GetContent()
		} else if len(params) == 1 {
			return template.Default(ctx).Label().
				SetContent(template.HTML(value.Value)).
				SetColor(params[0].Color).
				SetType(params[0].Type).
				GetContent()
		}
		return ""
	}
}
