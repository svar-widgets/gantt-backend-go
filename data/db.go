package data

import (
	"log"

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

	Tasks *TasksDAO
	Links *LinksDAO
}

func (d *DAO) GetDB() *gorm.DB {
	return d.db
}

func (d *DAO) mustExec(sql string) {
	err := d.db.Exec(sql).Error
	if err != nil {
		panic(err)
	}
}

func NewDAO(config DBConfig, url string) *DAO {
	db, err := gorm.Open(sqlite.Open(config.Path), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Error),
	})
	if err != nil {
		panic("failed to connect database")
	}

	db.AutoMigrate(&Task{})
	db.AutoMigrate(&Link{})

	dao := &DAO{
		db:    db,
		Tasks: NewTasksDAO(db),
		Links: NewLinksDAO(db),
	}

	if config.ResetOnStart {
		dataDown(dao)
		dataUp(dao)
	}

	return dao
}
