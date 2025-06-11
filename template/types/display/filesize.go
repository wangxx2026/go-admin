package display

import (
	"strconv"

	"github.com/wangxx2026/go-admin/context"
	"github.com/wangxx2026/go-admin/modules/utils"
	"github.com/wangxx2026/go-admin/template/types"
)

type FileSize struct {
	types.BaseDisplayFnGenerator
}

func init() {
	types.RegisterDisplayFnGenerator("filesize", new(FileSize))
}

func (f *FileSize) Get(ctx *context.Context, args ...interface{}) types.FieldFilterFn {
	return func(value types.FieldModel) interface{} {
		size, _ := strconv.ParseUint(value.Value, 10, 64)
		return utils.FileSize(size)
	}
}
