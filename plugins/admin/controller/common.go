package controller

import (
	"bytes"
	template2 "html/template"
	"net/http"
	"regexp"
	"strings"
	"sync"

	"github.com/wangxx2026/go-admin/template/types/action"

	"github.com/wangxx2026/go-admin/context"
	"github.com/wangxx2026/go-admin/modules/auth"
	c "github.com/wangxx2026/go-admin/modules/config"
	"github.com/wangxx2026/go-admin/modules/db"
	"github.com/wangxx2026/go-admin/modules/language"
	"github.com/wangxx2026/go-admin/modules/menu"
	"github.com/wangxx2026/go-admin/modules/service"
	"github.com/wangxx2026/go-admin/plugins/admin/models"
	"github.com/wangxx2026/go-admin/plugins/admin/modules/constant"
	"github.com/wangxx2026/go-admin/plugins/admin/modules/form"
	"github.com/wangxx2026/go-admin/plugins/admin/modules/table"
	"github.com/wangxx2026/go-admin/template"
	"github.com/wangxx2026/go-admin/template/icon"
	"github.com/wangxx2026/go-admin/template/types"
)

type Handler struct {
	config        *c.Config
	captchaConfig map[string]string
	services      service.List
	conn          db.Connection
	routes        context.RouterMap
	generators    table.GeneratorList
	operations    []context.Node
	navButtons    *types.Buttons
	operationLock sync.Mutex
	assetsTheme   map[string]string
}

func New(cfg ...Config) *Handler {
	if len(cfg) == 0 {
		return &Handler{
			operations: make([]context.Node, 0),
			navButtons: new(types.Buttons),
		}
	}
	return &Handler{
		config:      cfg[0].Config,
		services:    cfg[0].Services,
		conn:        cfg[0].Connection,
		generators:  cfg[0].Generators,
		operations:  make([]context.Node, 0),
		navButtons:  new(types.Buttons),
		assetsTheme: make(map[string]string),
	}
}

type Config struct {
	Config     *c.Config
	Services   service.List
	Connection db.Connection
	Generators table.GeneratorList
}

func (h *Handler) UpdateCfg(cfg Config) {
	h.config = cfg.Config
	h.services = cfg.Services
	h.conn = cfg.Connection
	h.generators = cfg.Generators
	h.assetsTheme = make(map[string]string)
}

func (h *Handler) SetCaptcha(captcha map[string]string) {
	h.captchaConfig = captcha
}

func (h *Handler) AssetsTheme(asset, theme string) {
	h.assetsTheme[asset] = theme
}

func (h *Handler) SetRoutes(r context.RouterMap) {
	h.routes = r
}

func (h *Handler) table(prefix string, ctx *context.Context) table.Table {
	t := h.generators[prefix](ctx)
	authHandler := auth.Middleware(db.GetConnection(h.services))
	for _, cb := range t.GetInfo().Callbacks {
		if cb.Value[constant.ContextNodeNeedAuth] == 1 {
			h.AddOperation(context.Node{
				Path:     cb.Path,
				Method:   cb.Method,
				Handlers: append([]context.Handler{authHandler}, cb.Handlers...),
			})
		} else {
			h.AddOperation(context.Node{Path: cb.Path, Method: cb.Method, Handlers: cb.Handlers})
		}
	}
	for _, cb := range t.GetForm().Callbacks {
		if cb.Value[constant.ContextNodeNeedAuth] == 1 {
			h.AddOperation(context.Node{
				Path:     cb.Path,
				Method:   cb.Method,
				Handlers: append([]context.Handler{authHandler}, cb.Handlers...),
			})
		} else {
			h.AddOperation(context.Node{Path: cb.Path, Method: cb.Method, Handlers: cb.Handlers})
		}
	}
	return t
}

func (h *Handler) route(name string) context.Router {
	return h.routes.Get(name)
}

func (h *Handler) routePath(name string, value ...string) string {
	return h.routes.Get(name).GetURL(value...)
}

func (h *Handler) routePathWithPrefix(name string, prefix string) string {
	return h.routePath(name, "prefix", prefix)
}

func (h *Handler) AddOperation(nodes ...context.Node) {
	h.operationLock.Lock()
	defer h.operationLock.Unlock()
	// TODO: 避免重复增加，第一次加入后，后面大部分会存在重复情况，以下循环可以优化
	addNodes := make([]context.Node, 0)
	for _, node := range nodes {
		if h.searchOperation(node.Path, node.Method) {
			continue
		}
		addNodes = append(addNodes, node)
	}
	h.operations = append(h.operations, addNodes...)
}

func (h *Handler) AddNavButton(btns *types.Buttons) {
	h.navButtons = btns
	for _, btn := range *btns {
		h.AddOperation(btn.GetAction().GetCallbacks())
	}
}

func (h *Handler) searchOperation(path, method string) bool {
	for _, node := range h.operations {
		if node.Path == path && node.Method == method {
			return true
		}
	}
	return false
}

func (h *Handler) OperationHandler(path string, ctx *context.Context) bool {
	for _, node := range h.operations {
		if node.Path == path {
			for _, handler := range node.Handlers {
				handler(ctx)
			}
			return true
		}
	}
	return false
}

func (h *Handler) HTML(ctx *context.Context, user models.UserModel, panel types.Panel,
	options ...template.ExecuteOptions) {
	buf := h.Execute(ctx, user, panel, "", options...)
	ctx.HTML(http.StatusOK, buf.String())
}

func (h *Handler) HTMLPlug(ctx *context.Context, user models.UserModel, panel types.Panel, plugName string,
	options ...template.ExecuteOptions) {
	var btns types.Buttons
	if plugName == "" {
		btns = (*h.navButtons).CheckPermission(user)
	} else {
		btns = (*h.navButtons).Copy().
			RemoveToolNavButton().
			RemoveSiteNavButton().
			RemoveInfoNavButton().
			Add(types.GetDropDownButton("", icon.Gear, []*types.NavDropDownItemButton{
				types.GetDropDownItemButton(language.GetFromHtml("plugin setting"),
					action.Jump(h.config.Url("/info/plugin_"+plugName+"/edit"))),
				types.GetDropDownItemButton(language.GetFromHtml("menus manage"),
					action.Jump(h.config.Url("/menu?__plugin_name="+plugName))),
			})).
			CheckPermission(user)
	}
	buf := h.ExecuteWithBtns(ctx, user, panel, plugName, btns, options...)
	ctx.HTML(http.StatusOK, buf.String())
}

func (h *Handler) ExecuteWithBtns(ctx *context.Context, user models.UserModel, panel types.Panel, plugName string, btns types.Buttons,
	options ...template.ExecuteOptions) *bytes.Buffer {

	tmpl, tmplName := aTemplate(ctx).GetTemplate(isPjax(ctx))
	option := template.GetExecuteOptions(options)

	return template.Execute(ctx, &template.ExecuteParam{
		User:       user,
		TmplName:   tmplName,
		Tmpl:       tmpl,
		Panel:      panel,
		Config:     h.config,
		Menu:       menu.GetGlobalMenu(user, h.conn, ctx.Lang(), plugName).SetActiveClass(h.config.URLRemovePrefix(ctx.Path())),
		Animation:  option.Animation,
		Buttons:    btns,
		Iframe:     ctx.IsIframe(),
		IsPjax:     isPjax(ctx),
		NoCompress: option.NoCompress,
	})
}

func (h *Handler) Execute(ctx *context.Context, user models.UserModel, panel types.Panel, plugName string,
	options ...template.ExecuteOptions) *bytes.Buffer {

	tmpl, tmplName := aTemplate(ctx).GetTemplate(isPjax(ctx))
	option := template.GetExecuteOptions(options)

	return template.Execute(ctx, &template.ExecuteParam{
		User:       user,
		TmplName:   tmplName,
		Tmpl:       tmpl,
		Panel:      panel,
		Config:     h.config,
		Menu:       menu.GetGlobalMenu(user, h.conn, ctx.Lang(), plugName).SetActiveClass(h.config.URLRemovePrefix(ctx.Path())),
		Animation:  option.Animation,
		Buttons:    (*h.navButtons).CheckPermission(user),
		Iframe:     ctx.IsIframe(),
		IsPjax:     isPjax(ctx),
		NoCompress: option.NoCompress,
	})
}

func isInfoUrl(s string) bool {
	reg, _ := regexp.Compile("(.*?)info/(.*?)$")
	sub := reg.FindStringSubmatch(s)
	return len(sub) > 2 && !strings.Contains(sub[2], "/")
}

func isNewUrl(s string, p string) bool {
	reg, _ := regexp.Compile("(.*?)info/" + p + "/new")
	return reg.MatchString(s)
}

func isEditUrl(s string, p string) bool {
	reg, _ := regexp.Compile("(.*?)info/" + p + "/edit")
	return reg.MatchString(s)
}

func (h *Handler) authSrv() *auth.TokenService {
	return auth.GetTokenService(h.services.Get(auth.TokenServiceKey))
}

func aAlert(ctx *context.Context) types.AlertAttribute {
	return aTemplate(ctx).Alert()
}

func aForm(ctx *context.Context) types.FormAttribute {
	return aTemplate(ctx).Form()
}

func aRow(ctx *context.Context) types.RowAttribute {
	return aTemplate(ctx).Row()
}

func aCol(ctx *context.Context) types.ColAttribute {
	return aTemplate(ctx).Col()
}

func aImage(ctx *context.Context) types.ImgAttribute {
	return aTemplate(ctx).Image()
}

func aButton(ctx *context.Context) types.ButtonAttribute {
	return aTemplate(ctx).Button()
}

func aTree(ctx *context.Context) types.TreeAttribute {
	return aTemplate(ctx).Tree()
}

func aTable(ctx *context.Context) types.TableAttribute {
	return aTemplate(ctx).Table()
}

func aDataTable(ctx *context.Context) types.DataTableAttribute {
	return aTemplate(ctx).DataTable()
}

func aBox(ctx *context.Context) types.BoxAttribute {
	return aTemplate(ctx).Box()
}

func aTab(ctx *context.Context) types.TabsAttribute {
	return aTemplate(ctx).Tabs()
}

func aTemplate(ctx *context.Context) template.Template {
	return template.Get(ctx, c.GetTheme())
}

func aTemplateByTheme(ctx *context.Context, theme string) template.Template {
	return template.Get(ctx, theme)
}

func isPjax(ctx *context.Context) bool {
	return ctx.IsPjax()
}

func formFooter(ctx *context.Context, page string, isHideEdit, isHideNew, isHideReset bool, btnWord template2.HTML) template2.HTML {
	col1 := aCol(ctx).SetSize(types.SizeMD(2)).GetContent()

	var (
		checkBoxs  template2.HTML
		checkBoxJS template2.HTML

		editCheckBox = template.HTML(`
			<label class="pull-right" style="margin: 5px 10px 0 0;">
                <input type="checkbox" class="continue_edit" style="position: absolute; opacity: 0;"> ` + language.Get("continue editing") + `
            </label>`)
		newCheckBox = template.HTML(`
			<label class="pull-right" style="margin: 5px 10px 0 0;">
                <input type="checkbox" class="continue_new" style="position: absolute; opacity: 0;"> ` + language.Get("continue creating") + `
            </label>`)

		editWithNewCheckBoxJs = template.HTML(`$('.continue_edit').iCheck({checkboxClass: 'icheckbox_minimal-blue'}).on('ifChanged', function (event) {
		if (this.checked) {
			$('.continue_new').iCheck('uncheck');
			$('input[name="` + form.PreviousKey + `"]').val(location.href)
		} else {
			$('input[name="` + form.PreviousKey + `"]').val(previous_url_goadmin)
		}
	});	`)

		newWithEditCheckBoxJs = template.HTML(`$('.continue_new').iCheck({checkboxClass: 'icheckbox_minimal-blue'}).on('ifChanged', function (event) {
		if (this.checked) {
			$('.continue_edit').iCheck('uncheck');
			$('input[name="` + form.PreviousKey + `"]').val(location.href.replace('/edit', '/new'))
		} else {
			$('input[name="` + form.PreviousKey + `"]').val(previous_url_goadmin)
		}
	});`)
	)

	if page == "edit" {
		if isHideNew {
			newCheckBox = ""
			newWithEditCheckBoxJs = ""
		}
		if isHideEdit {
			editCheckBox = ""
			editWithNewCheckBoxJs = ""
		}
		checkBoxs = editCheckBox + newCheckBox
		checkBoxJS = `<script>	
	let previous_url_goadmin = $('input[name="` + form.PreviousKey + `"]').attr("value")
	` + editWithNewCheckBoxJs + newWithEditCheckBoxJs + `
</script>
`
	} else if page == "edit_only" && !isHideEdit {
		checkBoxs = editCheckBox
		checkBoxJS = template.HTML(`	<script>
	let previous_url_goadmin = $('input[name="` + form.PreviousKey + `"]').attr("value")
	$('.continue_edit').iCheck({checkboxClass: 'icheckbox_minimal-blue'}).on('ifChanged', function (event) {
		if (this.checked) {
			$('input[name="` + form.PreviousKey + `"]').val(location.href)
		} else {
			$('input[name="` + form.PreviousKey + `"]').val(previous_url_goadmin)
		}
	});
</script>
`)
	} else if page == "new" && !isHideNew {
		checkBoxs = newCheckBox
		checkBoxJS = template.HTML(`	<script>
	let previous_url_goadmin = $('input[name="` + form.PreviousKey + `"]').attr("value")
	$('.continue_new').iCheck({checkboxClass: 'icheckbox_minimal-blue'}).on('ifChanged', function (event) {
		if (this.checked) {
			$('input[name="` + form.PreviousKey + `"]').val(location.href)
		} else {
			$('input[name="` + form.PreviousKey + `"]').val(previous_url_goadmin)
		}
	});
</script>
`)
	}

	btn1 := aButton(ctx).
		SetType("submit").
		AddClass("submit").
		SetContent(btnWord).
		SetThemePrimary().
		SetOrientationRight().
		GetContent()
	btn2 := template.HTML("")
	if !isHideReset {
		btn2 = aButton(ctx).
			SetType("reset").
			AddClass("reset").
			SetContent(language.GetFromHtml("Reset")).
			SetThemeWarning().
			SetOrientationLeft().
			GetContent()
	}
	col2 := aCol(ctx).SetSize(types.SizeMD(8)).
		SetContent(btn1 + checkBoxs + btn2 + checkBoxJS).GetContent()
	return col1 + col2
}

func filterFormFooter(ctx *context.Context, infoUrl string) template2.HTML {
	col1 := aCol(ctx).SetSize(types.SizeMD(2)).GetContent()
	btn1 := aButton(ctx).SetType("submit").
		AddClass("submit").
		SetContent(icon.Icon(icon.Search, 2) + language.GetFromHtml("search")).
		SetThemePrimary().
		SetSmallSize().
		SetOrientationLeft().
		SetLoadingText(icon.Icon(icon.Spinner, 1) + language.GetFromHtml("search")).
		GetContent()
	btn2 := aButton(ctx).SetType("reset").
		AddClass("reset").
		SetContent(icon.Icon(icon.Undo, 2) + language.GetFromHtml("reset")).
		SetThemeDefault().
		SetOrientationLeft().
		SetSmallSize().
		SetHref(infoUrl).
		SetMarginLeft(12).
		GetContent()
	col2 := aCol(ctx).SetSize(types.SizeMD(8)).
		SetContent(btn1 + btn2).GetContent()
	return col1 + col2
}

func formContent(ctx *context.Context, form types.FormAttribute, isTab, iframe, isHideBack bool, header template2.HTML) template2.HTML {
	if isTab {
		return form.GetContent()
	}
	if iframe {
		header = ""
	} else if header == template2.HTML("") {
		header = form.GetDefaultBoxHeader(isHideBack)
	}
	return aBox(ctx).
		SetHeader(header).
		WithHeadBorder().
		SetStyle(" ").
		SetIframeStyle(iframe).
		SetBody(form.GetContent()).
		GetContent()
}

func detailContent(ctx *context.Context, form types.FormAttribute, editUrl, deleteUrl string, iframe bool) template2.HTML {
	return aBox(ctx).
		SetHeader(form.GetDetailBoxHeader(editUrl, deleteUrl)).
		WithHeadBorder().
		SetBody(form.GetContent()).
		SetIframeStyle(iframe).
		GetContent()
}

func menuFormContent(ctx *context.Context, form types.FormAttribute) template2.HTML {
	return aBox(ctx).
		SetHeader(form.GetBoxHeaderNoButton()).
		SetStyle(" ").
		WithHeadBorder().
		SetBody(form.GetContent()).
		GetContent()
}
