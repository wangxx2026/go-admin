package db

import (
	"database/sql"
	"errors"
	"gorm.io/gorm"
	"sync"

	"github.com/wangxx2026/go-admin/modules/config"
	"xorm.io/xorm"
)

// Base is a common Connection.
type Base struct {
	DbList   map[string]*sql.DB
	GormList map[string]*gorm.DB
	Once     sync.Once
	Configs  config.DatabaseList
}

// Close implements the method Connection.Close.
func (db *Base) Close() []error {
	errs := make([]error, 0)
	for _, d := range db.DbList {
		errs = append(errs, d.Close())
	}
	return errs
}

// GetDB implements the method Connection.GetDB.
func (db *Base) GetDB(key string) *sql.DB {
	return db.DbList[key]
}
func (db *Base) GetGormDB(key string) *gorm.DB {
	return db.GormList[key]
}

func (db *Base) CreateDB(name string, beans ...interface{}) error {
	cfg := db.GetConfig(name)
	if cfg.Driver == "" {
		return errors.New("wrong connection name")
	}
	engine, err := xorm.NewEngine(cfg.Driver, cfg.GetDSN())
	if err != nil {
		return err
	}
	defer func() {
		_ = engine.Close()
	}()
	err = engine.Sync(beans...)
	if err != nil {
		return err
	}
	return nil
}

func (db *Base) GetConfig(name string) config.Database {
	return db.Configs[name]
}
