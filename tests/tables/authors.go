package tables

import (
	"github.com/wangxx2026/go-admin/context"
	"github.com/wangxx2026/go-admin/modules/db"
	"github.com/wangxx2026/go-admin/plugins/admin/modules/table"
	"github.com/wangxx2026/go-admin/template/types/form"
)

// GetAuthorsTable return the model of table author.
func GetAuthorsTable(ctx *context.Context) (authorsTable table.Table) {

	authorsTable = table.NewDefaultTable(ctx, table.DefaultConfig())

	// connect your custom connection
	// authorsTable = table.NewDefaultTable(ctx, table.DefaultConfigWithDriverAndConnection("mysql", "admin"))

	info := authorsTable.GetInfo()
	info.AddField("ID", "id", db.Int).FieldSortable()
	info.AddField("First Name", "first_name", db.Varchar)
	info.AddField("Last Name", "last_name", db.Varchar)
	info.AddField("Email", "email", db.Varchar)
	info.AddField("Birthdate", "birthdate", db.Date)
	info.AddField("Added", "added", db.Timestamp)

	info.SetTable("authors").SetTitle("Authors").SetDescription("Authors")

	formList := authorsTable.GetForm()
	formList.AddField("ID", "id", db.Int, form.Default).FieldDisplayButCanNotEditWhenUpdate().FieldDisableWhenCreate()
	formList.AddField("First Name", "first_name", db.Varchar, form.Text)
	formList.AddField("Last Name", "last_name", db.Varchar, form.Text)
	formList.AddField("Email", "email", db.Varchar, form.Text)
	formList.AddField("Birthdate", "birthdate", db.Date, form.Text)
	formList.AddField("Added", "added", db.Timestamp, form.Text)

	formList.SetTable("authors").SetTitle("Authors").SetDescription("Authors")

	return
}
