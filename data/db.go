package data

import (
	"log"
	"strings"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var Debug = 1

func logError(e error) {
	if e != nil && Debug > 0 {
		log.Printf("[ERROR]\n%s\n", e)
	}
}

type DBConfig struct {
	Path         string
	ResetOnStart bool
}

type DAO struct {
	db *gorm.DB

	Tasks       *TasksDAO
	Links       *LinksDAO
	Resources   *ResourcesDAO
	Assignments *AssignmentsDAO
}

func (d *DAO) GetDB() *gorm.DB {
	return d.db
}

func (d *DAO) withTx(tx *gorm.DB) *DAO {
	return &DAO{
		db:          tx,
		Tasks:       NewTasksDAO(tx),
		Links:       NewLinksDAO(tx),
		Resources:   NewResourcesDAO(tx),
		Assignments: NewAssignmentsDAO(tx),
	}
}

func (d *DAO) Transaction(fn func(dao *DAO) error) error {
	return d.db.Transaction(func(tx *gorm.DB) error {
		return fn(d.withTx(tx))
	})
}

func (d *DAO) mustExec(sql string) {
	err := d.db.Exec(sql).Error
	if err != nil {
		panic(err)
	}
}

func NewDAO(config DBConfig, url string) *DAO {
	dsn := config.Path
	sep := "?"
	if strings.Contains(dsn, "?") {
		sep = "&"
	}
	dsn += sep + "_busy_timeout=5000"

	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Error),
	})
	if err != nil {
		panic("failed to connect database")
	}

	if sqlDB, err := db.DB(); err == nil {
		sqlDB.SetMaxOpenConns(1)
	}

	db.AutoMigrate(&Task{})
	db.AutoMigrate(&Link{})
	db.AutoMigrate(&Resource{})
	db.AutoMigrate(&Assignment{})

	dao := &DAO{
		db:          db,
		Tasks:       NewTasksDAO(db),
		Links:       NewLinksDAO(db),
		Resources:   NewResourcesDAO(db),
		Assignments: NewAssignmentsDAO(db),
	}

	if config.ResetOnStart {
		dataDown(dao)
		dataUp(dao)
	}

	return dao
}
