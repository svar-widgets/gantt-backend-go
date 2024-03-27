package data

import (
	"encoding/json"
	"log"
	"os"
)

func dataDown(d *DAO) {
	d.mustExec("DELETE from tasks")
	d.mustExec("DELETE from links")
}

func dataUp(d *DAO) (err error) {
	tasks := make([]Task, 0)
	err = parseDemodata(&tasks, "./demodata/tasks.json")
	if err != nil {
		log.Fatal(err)
	}
	indexPull := make(map[int]int)
	for i := range tasks {
		tasks[i].Index = indexPull[tasks[i].Parent]
		indexPull[tasks[i].Parent]++
	}

	links := make([]Link, 0)
	err = parseDemodata(&links, "./demodata/links.json")
	if err != nil {
		log.Fatal(err)
	}

	db := d.GetDB()
	err = db.Create(&tasks).Error
	if err != nil {
		return err
	}
	err = db.Create(&links).Error

	return
}

func parseDemodata(dest interface{}, path string) error {
	bytes, err := os.ReadFile(path)
	if err != nil {
		logError(err)
		return err
	}
	err = json.Unmarshal(bytes, &dest)
	if err != nil {
		logError(err)
	}
	return err
}
