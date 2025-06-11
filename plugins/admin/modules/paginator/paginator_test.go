package paginator

import (
	"testing"

	_ "github.com/GoAdminGroup/themes/sword"
	"github.com/wangxx2026/go-admin/modules/config"
	"github.com/wangxx2026/go-admin/plugins/admin/modules/parameter"
)

func TestGet(t *testing.T) {
	config.Initialize(&config.Config{Theme: "sword"})
	param := parameter.BaseParam()
	param.Page = "7"
	Get(nil, Config{
		Size:         105,
		Param:        param,
		PageSizeList: []string{"10", "20", "50", "100"},
	})
}
