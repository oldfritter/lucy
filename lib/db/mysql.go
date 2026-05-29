package db

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"

	"github.com/oldfritter/lucy/base"
	"github.com/oldfritter/lucy/util"
)

var (
	MysqlDB *gorm.DB
)

func init() {
	if MysqlDB == nil {
		MysqlDB = initDBConnect("mysql")
	}
}

func initDBConnect(name string) *gorm.DB {
	config := base.GetConfig()
	file := util.GetLogFile("gorm")
	level := logger.Info
	if config.Get("model", "") == "production" {
		level = logger.Error
	}
	newLogger := logger.New(
		log.New(file, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second * 2,
			LogLevel:                  level,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		},
	)
	var err error
	conn := fmt.Sprintf("%v:%v@tcp(%v:%v)/%v?charset=utf8mb4&parseTime=True&loc=Local&timeout=2000ms",
		config.Get(name+".username", ""),
		config.Get(name+".password", ""),
		config.Get(name+".host", ""),
		config.Get(name+".port", ""),
		config.Get(name+".database", ""),
	)
	con, err := gorm.Open(
		mysql.Open(conn),
		&gorm.Config{
			DisableForeignKeyConstraintWhenMigrating: true,
			NamingStrategy: schema.NamingStrategy{
				SingularTable: true,
			},
			Logger: newLogger,
		})
	if err != nil {
		log.Fatalf("database %s connect error %v", config.Get(name+".database", ""), err)
	}
	db, err := con.DB()
	if err != nil {
		log.Fatalf("database %s connect error %v", config.Get(name+".database", ""), err)
	} else {
		log.Printf("database %s connect success!", config.Get(name+".database", ""))
	}
	db.SetMaxIdleConns(10)
	db.SetConnMaxIdleTime(10 * time.Minute)
	db.SetMaxOpenConns(config.GetInt(name+".pool", 10))
	if con.Error != nil {
		log.Fatalf("database error %v", con.Error)
	}
	return con
}

func CloseDB() {
	if dbSQL, ok := MysqlDB.DB(); ok != nil {
		defer dbSQL.Close()
	}
}

type GormDB struct {
	*gorm.DB
	gdbDone bool
}

func BeginTx() *GormDB {
	txn := MysqlDB.Begin()
	if txn.Error != nil {
		panic(txn.Error)
	}
	return &GormDB{txn, false}
}

func (c *GormDB) DbCommit() {
	if c.gdbDone {
		return
	}
	tx := c.Commit()
	c.gdbDone = true
	if err := tx.Error; err != nil && err != sql.ErrTxDone {
		panic(err)
	}
}

func (c *GormDB) DbRollback() {
	if c.gdbDone {
		return
	}
	tx := c.Rollback()
	c.gdbDone = true
	if err := tx.Error; err != nil && err != sql.ErrTxDone {
		panic(err)
	}
}

func (c *GormDB) Error() error {
	return c.DB.Error
}
