package main

import (
	"log"
	"os"
	"os/signal"
	"time"

	_ "github.com/wangxx2026/go-admin/adapter/echo"
	_ "github.com/wangxx2026/go-admin/modules/db/drivers/mysql"
	_ "github.com/wangxx2026/themes/adminlte"

	"github.com/labstack/echo/v4"
	"github.com/wangxx2026/go-admin/engine"
	"github.com/wangxx2026/go-admin/examples/datamodel"
	"github.com/wangxx2026/go-admin/modules/config"
	"github.com/wangxx2026/go-admin/modules/language"
	"github.com/wangxx2026/go-admin/plugins/example"
	"github.com/wangxx2026/go-admin/template"
	"github.com/wangxx2026/go-admin/template/chartjs"
)

func main() {
	e := echo.New()

	eng := engine.Default()

	cfg := config.Config{
		Env: config.EnvLocal,
		Databases: config.DatabaseList{
			"default": {
				Host:            "127.0.0.1",
				Port:            "3306",
				User:            "root",
				Pwd:             "root",
				Name:            "godmin",
				MaxIdleConns:    50,
				MaxOpenConns:    150,
				ConnMaxLifetime: time.Hour,
				Driver:          config.DriverMysql,
			},
		},
		UrlPrefix: "admin",
		IndexUrl:  "/",
		Store: config.Store{
			Path:   "./uploads",
			Prefix: "uploads",
		},
		Debug:    true,
		Language: language.CN,
	}

	template.AddComp(chartjs.NewChart())

	// customize a plugin

	examplePlugin := example.NewExample()

	// load from golang.Plugin
	//
	// examplePlugin := plugins.LoadFromPlugin("../datamodel/example.so")

	// customize the login page
	// example: https://github.com/GoAdminGroup/demo.go-admin.cn/blob/master/main.go#L39
	//
	// template.AddComp("login", datamodel.LoginPage)

	// load config from json file
	//
	// eng.AddConfigFromJSON("../datamodel/config.json")

	if err := eng.AddConfig(&cfg).
		AddGenerators(datamodel.Generators).
		AddDisplayFilterXssJsFilter().
		// add generator, first parameter is the url prefix of table when visit.
		// example:
		//
		// "user" => http://localhost:9033/admin/info/user
		//
		AddGenerator("user", datamodel.GetUserTable).
		AddPlugins(examplePlugin).
		Use(e); err != nil {
		panic(err)
	}

	e.Static("/uploads", "./uploads")

	// you can custom your pages like:

	eng.HTML("GET", "/admin", datamodel.GetContent)

	// Start server
	go e.Logger.Fatal(e.Start(":1323"))

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	log.Print("closing database connection")
	eng.MysqlConnection().Close()
}
