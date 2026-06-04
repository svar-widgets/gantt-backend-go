package data

import (
	"fmt"

	"gorm.io/gorm"
)

type AssignmentsDAO struct {
	db *gorm.DB
}

type AssignmentPayload struct {
	Resource int `json:"resource"`
	Units    int `json:"units"`
	Task     int `json:"task"`
}

func NewAssignmentsDAO(db *gorm.DB) *AssignmentsDAO {
	return &AssignmentsDAO{db}
}

func (d *AssignmentsDAO) GetAll() ([]Assignment, error) {
	assignments := make([]Assignment, 0)
	err := d.db.Order("id").Find(&assignments).Error
	return assignments, err
}

func (d *AssignmentsDAO) GetByTasks(taskIDs []int) ([]Assignment, error) {
	assignments := make([]Assignment, 0)
	if len(taskIDs) == 0 {
		return assignments, nil
	}
	err := d.db.Where("task IN ?", taskIDs).Order("id").Find(&assignments).Error
	return assignments, err
}

func (d *AssignmentsDAO) GetOne(id int) (Assignment, error) {
	assignment := Assignment{}
	err := d.db.Find(&assignment, id).Error
	if assignment.ID == 0 {
		return Assignment{}, fmt.Errorf("assignment with id %d not found", id)
	}
	return assignment, err
}

func (d *AssignmentsDAO) Add(data AssignmentPayload) (int, error) {
	assignment := Assignment{}
	data.fillModel(&assignment)
	err := d.db.Create(&assignment).Error
	
	return assignment.ID, err
}

func (d *AssignmentsDAO) Update(id int, data AssignmentPayload) error {
	assignment, err := d.GetOne(id)
	if err != nil {
		return err
	}

	data.fillModel(&assignment)
	return d.db.Save(&assignment).Error
}

func (d *AssignmentsDAO) Delete(id int) error {
	if _, err := d.GetOne(id); err != nil {
		return err
	}
	return d.db.Delete(&Assignment{}, id).Error
}

func (d *AssignmentsDAO) DeleteByTasks(taskIDs []int) error {
	if len(taskIDs) == 0 {
		return nil
	}
	return d.db.Where("task IN ?", taskIDs).Delete(&Assignment{}).Error
}
func (u *AssignmentPayload) fillModel(model *Assignment) {
	model.Resource = u.Resource
	model.Units = u.Units
	model.Task = u.Task
}
