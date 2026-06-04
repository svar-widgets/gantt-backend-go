package data

import (
	"fmt"

	"gorm.io/gorm"
)

type ResourcesDAO struct {
	db *gorm.DB
}

func NewResourcesDAO(db *gorm.DB) *ResourcesDAO {
	return &ResourcesDAO{db}
}

func (d *ResourcesDAO) GetAll() ([]Resource, error) {
	resources := make([]Resource, 0)
	err := d.db.Order("parent, id").Find(&resources).Error
	return resources, err
}

func (d *ResourcesDAO) GetOne(id int) (Resource, error) {
	resource := Resource{}
	err := d.db.Find(&resource, id).Error
	if resource.ID == 0 {
		return Resource{}, fmt.Errorf("resource with id %d not found", id)
	}
	return resource, err
}
