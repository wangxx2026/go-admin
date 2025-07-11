package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/wangxx2026/go-admin/adapter/nethttp"
	_ "github.com/wangxx2026/go-admin/modules/db/drivers/mysql"

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
	r := http.NewServeMux()

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
		Store: config.Store{
			Path:   "./uploads",
			Prefix: "uploads",
		},
		Language:    language.EN,
		IndexUrl:    "/",
		Debug:       true,
		ColorScheme: adminlte.ColorschemeSkinBlack,
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
		Use(r); err != nil {
		panic(err)
	}

	workDir, _ := os.Getwd()
	filesDir := filepath.Join(workDir, "uploads")
	FileServer(r, "/uploads", http.Dir(filesDir))

	// you can custom your pages like:

	eng.HTML("GET", "/admin", datamodel.GetContent)

	go func() {
		_ = http.ListenAndServe(":3333", r)
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	log.Print("closing database connection")
	eng.MysqlConnection().Close()
}

// FileServer conveniently sets up a http.FileServer handler to serve
// static files from a http.FileSystem.
func FileServer(r *http.ServeMux, path string, root http.FileSystem) {
	if strings.ContainsAny(path, "{}*") {
		panic("FileServer does not permit URL parameters.")
	}

	fs := http.StripPrefix(path, http.FileServer(root))

	if path != "/" && path[len(path)-1] != '/' {
		r.HandleFunc("GET "+path, http.RedirectHandler(path+"/", http.StatusMovedPermanently).ServeHTTP)
		path += "/"
	}
	path += "*"

	r.HandleFunc("GET "+path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fs.ServeHTTP(w, r)
	}))
}
