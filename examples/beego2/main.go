package main

import (
	"log"
	"os"
	"os/signal"
	"time"

	_ "github.com/wangxx2026/go-admin/adapter/beego2"
	_ "github.com/wangxx2026/go-admin/modules/db/drivers/mysql"

	"github.com/beego/beego/v2/server/web"
	"github.com/wangxx2026/go-admin/engine"
	"github.com/wangxx2026/go-admin/examples/datamodel"
	"github.com/wangxx2026/go-admin/modules/config"
	"github.com/wangxx2026/go-admin/modules/language"
	"github.com/wangxx2026/go-admin/plugins/example"
	"github.com/wangxx2026/go-admin/template"
	"github.com/wangxx2026/go-admin/template/chartjs"
	"github.com/wangxx2026/themes/adminlte"
)

func main() {
	app := web.NewHttpSever()

	eng := engine.Default()

	cfg := config.Config{
		Env: config.EnvLocal,
		Databases: config.DatabaseList{
			"default": {
				Host:            "127.0.0.1",
				Port:            "3306",
				User:            "root",
				Pwd:             "123456",
				Name:            "godmin",
				MaxIdleConns:    50,
				MaxOpenConns:    150,
				ConnMaxLifetime: time.Hour,
				Driver:          config.DriverMysql,
			},
		},
		Store: config.Store{
			Path:   "./uploads",
			Prefix: "uploads",
		},
		UrlPrefix:   "admin",
		IndexUrl:    "/",
		Debug:       true,
		Language:    language.CN,
		ColorScheme: adminlte.ColorschemeSkinBlack,
	}

	template.AddComp(chartjs.NewChart())

	examplePlugin := example.NewExample()

	web.SetStaticPath("/uploads", "uploads")

	if err := eng.AddConfig(&cfg).
		AddGenerators(datamodel.Generators).
		AddDisplayFilterXssJsFilter().
		AddGenerator("user", datamodel.GetUserTable).
		AddPlugins(examplePlugin).
		Use(app); err != nil {
		panic(err)
	}

	eng.HTML("GET", "/admin", datamodel.GetContent)

	app.Cfg.Listen.HTTPSAddr = "127.0.0.1"
	app.Cfg.Listen.HTTPPort = 9087
	go app.Run("")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	log.Print("closing database connection")
	eng.MysqlConnection().Close()
}
