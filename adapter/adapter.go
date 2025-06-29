// Copyright 2019 GoAdmin Core Team. All rights reserved.
// Use of this source code is governed by a Apache-2.0 style
// license that can be found in the LICENSE file.

package adapter

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"

	"github.com/wangxx2026/go-admin/context"
	"github.com/wangxx2026/go-admin/modules/auth"
	"github.com/wangxx2026/go-admin/modules/config"
	"github.com/wangxx2026/go-admin/modules/constant"
	"github.com/wangxx2026/go-admin/modules/db"
	"github.com/wangxx2026/go-admin/modules/errors"
	"github.com/wangxx2026/go-admin/modules/logger"
	"github.com/wangxx2026/go-admin/modules/menu"
	"github.com/wangxx2026/go-admin/plugins"
	"github.com/wangxx2026/go-admin/plugins/admin/models"
	"github.com/wangxx2026/go-admin/template"
	"github.com/wangxx2026/go-admin/template/types"
)

// WebFrameWork is an interface which is used as an adapter of
// framework and goAdmin. It must implement two methods. Use registers
// the routes and the corresponding handlers. Content writes the
// response to the corresponding context of framework.
type WebFrameWork interface {
	// Name return the web framework name.
	Name() string

	// Use method inject the plugins to the web framework engine which is the
	// first parameter.
	Use(app interface{}, plugins []plugins.Plugin) error

	// Content add the panel html response of the given callback function to
	// the web framework context which is the first parameter.
	Content(ctx interface{}, fn types.GetPanelFn, fn2 context.NodeProcessor, navButtons ...types.Button)

	// User get the auth user model from the given web framework context.
	User(ctx interface{}) (models.UserModel, bool)

	// AddHandler inject the route and handlers of GoAdmin to the web framework.
	AddHandler(method, path string, handlers context.Handlers)

	DisableLog()

	Static(prefix, path string)

	Run() error

	// Helper functions
	// ================================

	SetApp(app interface{}) error
	SetConnection(db.Connection)
	GetConnection() db.Connection
	SetContext(ctx interface{}) WebFrameWork
	GetCookie() (string, error)
	Lang() string
	Path() string
	Method() string
	Request() *http.Request
	FormParam() url.Values
	Query() url.Values
	IsPjax() bool
	Redirect()
	SetContentType()
	Write(body []byte)
	CookieKey() string
	HTMLContentType() string
}

// BaseAdapter is a base adapter contains some helper functions.
type BaseAdapter struct {
	db db.Connection
}

// SetConnection set the db connection.
func (base *BaseAdapter) SetConnection(conn db.Connection) {
	base.db = conn
}

// GetConnection get the db connection.
func (base *BaseAdapter) GetConnection() db.Connection {
	return base.db
}

// HTMLContentType return the default content type header.
func (*BaseAdapter) HTMLContentType() string {
	return "text/html; charset=utf-8"
}

// CookieKey return the cookie key.
func (*BaseAdapter) CookieKey() string {
	return auth.DefaultCookieKey
}

// GetUser is a helper function get the auth user model from the context.
func (*BaseAdapter) GetUser(ctx interface{}, wf WebFrameWork) (models.UserModel, bool) {
	cookie, err := wf.SetContext(ctx).GetCookie()

	if err != nil {
		return models.UserModel{}, false
	}

	user, exist := auth.GetCurUser(cookie, wf.GetConnection())
	return user.ReleaseConn(), exist
}

// GetUse is a helper function adds the plugins to the framework.
func (*BaseAdapter) GetUse(app interface{}, plugin []plugins.Plugin, wf WebFrameWork) error {
	if err := wf.SetApp(app); err != nil {
		return err
	}

	for _, plug := range plugin {
		for path, handlers := range plug.GetHandler() {
			if plug.Prefix() == "" {
				wf.AddHandler(path.Method, path.URL, handlers)
			} else {
				wf.AddHandler(path.Method, config.Url("/"+plug.Prefix()+path.URL), handlers)
			}
		}
	}

	return nil
}

func (*BaseAdapter) Run() error         { panic("not implement") }
func (*BaseAdapter) DisableLog()        { panic("not implement") }
func (*BaseAdapter) Static(_, _ string) { panic("not implement") }

// GetContent is a helper function of adapter.Content
func (base *BaseAdapter) GetContent(ctx interface{}, getPanelFn types.GetPanelFn, wf WebFrameWork,
	navButtons types.Buttons, fn context.NodeProcessor) {

	var (
		newBase          = wf.SetContext(ctx)
		cookie, hasError = newBase.GetCookie()
	)

	if hasError != nil || cookie == "" {
		newBase.Redirect()
		return
	}

	user, authSuccess := auth.GetCurUser(cookie, wf.GetConnection())

	if !authSuccess {
		newBase.Redirect()
		return
	}

	var (
		panel types.Panel
		err   error
	)

	gctx := context.NewContext(newBase.Request())

	if !auth.CheckPermissions(user, newBase.Path(), newBase.Method(), newBase.FormParam()) {
		panel = template.WarningPanel(gctx, errors.NoPermission, template.NoPermission403Page)
	} else {
		panel, err = getPanelFn(ctx)
		if err != nil {
			panel = template.WarningPanel(gctx, err.Error())
		}
	}

	fn(panel.Callbacks...)

	tmpl, tmplName := template.Default(gctx).GetTemplate(newBase.IsPjax())

	buf := new(bytes.Buffer)
	hasError = tmpl.ExecuteTemplate(buf, tmplName, types.NewPage(gctx, &types.NewPageParam{
		User:         user,
		Menu:         menu.GetGlobalMenu(user, wf.GetConnection(), newBase.Lang()).SetActiveClass(config.URLRemovePrefix(newBase.Path())),
		Panel:        panel.GetContent(config.IsProductionEnvironment()),
		Assets:       template.GetComponentAssetImportHTML(gctx),
		Buttons:      navButtons.CheckPermission(user),
		TmplHeadHTML: template.Default(gctx).GetHeadHTML(),
		TmplFootJS:   template.Default(gctx).GetFootJS(),
		Iframe:       newBase.Query().Get(constant.IframeKey) == "true",
	}))

	if hasError != nil {
		logger.Error(fmt.Sprintf("error: %s adapter content, ", newBase.Name()), hasError)
	}

	newBase.SetContentType()
	newBase.Write(buf.Bytes())
}
