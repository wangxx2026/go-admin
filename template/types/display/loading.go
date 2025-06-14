package display

import (
	"html/template"

	"github.com/wangxx2026/go-admin/context"
	"github.com/wangxx2026/go-admin/template/types"
)

type Loading struct {
	types.BaseDisplayFnGenerator
}

func init() {
	types.RegisterDisplayFnGenerator("loading", new(Loading))
}

func (l *Loading) Get(ctx *context.Context, args ...interface{}) types.FieldFilterFn {
	return func(value types.FieldModel) interface{} {
		param := args[0].([]string)

		for i := 0; i < len(param); i++ {
			if value.Value == param[i] {
				return template.HTML(`<i class="fa fa-refresh fa-spin text-primary"></i>`)
			}
		}

		return value.Value
	}
}
