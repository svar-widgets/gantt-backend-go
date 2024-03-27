package data

import (
	"fmt"
	"gantt-backend-go/common"

	"gorm.io/gorm"
)

type TasksDAO struct {
	db *gorm.DB
}

type TaskUpdate struct {
	Text     string          `json:"text"`
	Start    *common.JDate   `json:"start"`
	End      *common.JDate   `json:"end"`
	Duration int             `json:"duration"`
	Progress int             `json:"progress"`
	Parent   common.FuzzyInt `json:"parent"`
	Type     string          `json:"type"`
	Lazy     bool            `json:"lazy"`
}

type UpdatePayload struct {
	TaskUpdate

	Operation string `json:"operation"`
	Target    int    `json:"target"`
	Mode      string `json:"mode"`
	Nested    bool   `json:"nested"`
}

func NewTasksDAO(db *gorm.DB) *TasksDAO {
	return &TasksDAO{db}
}

func (d *TasksDAO) GetOne(id int) (Task, error) {
	task := Task{}
	err := d.db.Find(&task, id).Error
	if task.ID == 0 {
		return Task{}, fmt.Errorf("task with id %d not found", id)
	}

	return task, err
}

func (d *TasksDAO) GetAll() ([]Task, error) {
	tasks := make([]Task, 0)
	err := d.db.Order("parent, `index`").Find(&tasks).Error

	return tasks, err
}

func (d *TasksDAO) GetBranch(id int) ([]Task, error) {
	return d.getBranch(nil, id)
}

func (d *TasksDAO) Add(data TaskUpdate) (int, error) {
	task := Task{}
	data.fillModel(&task)

	branch, err := d.getKids(nil, task.Parent)
	if err != nil {
		return task.ID, err
	}
	task.Index = len(branch)

	err = d.db.Create(&task).Error

	return task.ID, err
}

func (d *TasksDAO) Update(id int, data UpdatePayload) error {
	task, err := d.GetOne(id)
	if err != nil {
		return err
	}
	data.fillModel(&task)
	return d.db.Save(&task).Error
}

func (d *TasksDAO) Delete(id int) ([]int, error) {
	task, err := d.GetOne(id)
	if err != nil {
		return nil, err
	}
	tasks, err := d.getBranch(nil, id)
	if err != nil {
		return nil, err
	}
	toRemove := make([]int, 0)
	toRemove = append(toRemove, id)
	for _, t := range tasks {
		toRemove = append(toRemove, t.ID)
	}
	err = d.db.Where("id IN ?", toRemove).Delete(&Task{}).Error

	if task.Parent != 0 {
		kids, err := d.getBranch(nil, task.Parent)
		if err != nil {
			return nil, err
		}
		if len(kids) == 0 {
			err = d.db.Model(&Task{}).Where("id = ?", task.Parent).Update("lazy", false).Error
			if err != nil {
				return nil, err
			}
		}
	}

	return toRemove, err
}

func (d *TasksDAO) Move(id int, data UpdatePayload) (err error) {
	task, err := d.GetOne(id)
	if err != nil {
		return err
	}
	target, err := d.GetOne(data.Target)
	if err != nil {
		return err
	}

	tx := d.db.Begin()
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	var targetParent int
	if data.Mode == "child" {
		targetParent = target.ID
	} else {
		targetParent = target.Parent
	}

	targetBranch, err := d.getKids(tx, int(targetParent))
	if err != nil {
		return err
	}

	l := len(targetBranch)

	if data.Mode == "child" {
		if l > 0 {
			data.Mode = "before"
			target = targetBranch[0]
		}
	}

	otherBranch := task.Parent != target.Parent || l == 0
	if otherBranch {
		l++
	}

	branchUpd := make([]Task, l)
	ind := 0
	if data.Mode == "child" && len(targetBranch) == 0 {
		branchUpd[ind] = task
	} else {
		for _, t := range targetBranch {
			if t.ID == id {
				continue
			}

			if t.ID == target.ID {
				if data.Mode == "after" {
					branchUpd[ind] = t
					branchUpd[ind+1] = task
					break
				} else {
					branchUpd[ind] = task
					ind++
				}
			}
			branchUpd[ind] = t
			ind++
		}
	}

	err = d.refreshBranchOrder(tx, branchUpd)
	if err != nil {
		return err
	}

	if otherBranch {
		err = tx.Model(&Task{}).Where("id = ?", id).Update("parent", targetParent).Error
		if err != nil {
			return err
		}
		oldBranch, err := d.getKids(tx, task.Parent)
		if err != nil {
			return err
		}
		l := len(oldBranch)
		if l == 0 {
			err := tx.Model(&Task{}).Where("id = ?", task.Parent).Update("lazy", false).Error
			if err != nil {
				return err
			}
		} else if l > 1 {
			err = d.refreshBranchOrder(tx, oldBranch)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (d *TasksDAO) Copy(id int, data UpdatePayload) ([]int, []int, error) {
	task, err := d.GetOne(id)
	if err != nil {
		return nil, nil, err
	}
	target, err := d.GetOne(data.Target)
	if err != nil {
		return nil, nil, err
	}

	var targetParent int
	if data.Mode == "child" {
		targetParent = target.ID
	} else {
		targetParent = target.Parent
	}

	targetBranch, err := d.getKids(nil, int(targetParent))
	if err != nil {
		return nil, nil, err
	}

	l := len(targetBranch)
	if data.Mode == "child" {
		if l > 0 {
			data.Mode = "after"
			target = targetBranch[l-1]
		}
	}

	ids, nids, err := d.createCopy(nil, task, targetParent, data.Nested)
	if err != nil {
		return nil, nil, err
	}

	ntask, err := d.GetOne(nids[0])
	if err != nil {
		return nil, nil, err
	}

	tx := d.db.Begin()
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	l = l + 1

	branchUpd := make([]Task, l)
	ind := 0
	if data.Mode == "child" && l == 1 {
		branchUpd[ind] = ntask
	} else {
		for _, t := range targetBranch {
			if t.ID == target.ID {
				if data.Mode == "after" {
					branchUpd[ind] = t
					ind++
					branchUpd[ind] = ntask
					ind++
					continue
				} else {
					branchUpd[ind] = ntask
					ind++
				}
			}
			branchUpd[ind] = t
			ind++
		}
	}

	err = d.refreshBranchOrder(tx, branchUpd)
	if err != nil {
		return nil, nil, err
	}

	return ids, nids, nil
}

func (d *TasksDAO) createCopy(tx *gorm.DB, task Task, parent int, nested bool) ([]int, []int, error) {
	ids := make([]int, 0)
	nids := make([]int, 0)
	nid, err := d.Add(TaskUpdate{
		Text:     task.Text,
		Start:    task.Start,
		End:      task.End,
		Duration: task.Duration,
		Progress: task.Progress,
		Parent:   common.FuzzyInt(parent),
		Type:     task.Type,
		Lazy:     task.Lazy,
	})
	if err != nil {
		return nil, nil, err
	}
	ids = append(ids, task.ID)
	nids = append(nids, nid)

	if nested {
		kids, err := d.getKids(tx, task.ID)
		if err != nil {
			return nil, nil, err
		}
		for _, kid := range kids {
			oids, knids, err := d.createCopy(tx, kid, nid, nested)
			if err != nil {
				return nil, nil, err
			}
			ids = append(ids, oids...)
			nids = append(nids, knids...)
		}
	}

	return ids, nids, nil
}

func (d *TasksDAO) refreshBranchOrder(tx *gorm.DB, branch []Task) error {
	var err error
	for i, t := range branch {
		branch[i].Index = i
		err = tx.Model(&Task{}).Where("id = ?", t.ID).Update("index", i).Error
		if err != nil {
			break
		}
	}
	return err
}

func (d *TasksDAO) getKids(tx *gorm.DB, parent int) ([]Task, error) {
	if tx == nil {
		tx = new(gorm.DB)
		*tx = *d.db
	}
	tasks := make([]Task, 0)
	err := tx.Where("parent = ?", parent).Order("parent, `index`").Find(&tasks).Error
	if err != nil {
		return nil, err
	}
	return tasks, nil
}

func (d *TasksDAO) getBranch(tx *gorm.DB, parent int) ([]Task, error) {
	if tx == nil {
		tx = new(gorm.DB)
		*tx = *d.db
	}

	if parent != 0 {
		_, err := d.GetOne(parent)
		if err != nil {
			return nil, err
		}
	}

	tasks, err := d.collectChildren(parent)
	if err != nil {
		return nil, err
	}
	return tasks, nil
}

func (d *TasksDAO) collectChildren(parent int) ([]Task, error) {
	tasks := make([]Task, 0)
	kids, err := d.getKids(nil, int(parent))
	if err != nil {
		return nil, err
	}
	tasks = append(tasks, kids...)
	for _, k := range kids {
		if !k.Lazy {
			kk, err := d.collectChildren(k.ID)
			if err != nil {
				return nil, err
			}
			tasks = append(tasks, kk...)
		}
	}
	return tasks, nil
}

func (u *TaskUpdate) fillModel(model *Task) {
	model.Text = u.Text
	if u.Start != nil {
		ts := common.JDate(*u.Start)
		model.Start = &ts
	}
	if u.End != nil {
		te := common.JDate(*u.End)
		model.End = &te
	}
	model.Duration = u.Duration
	model.Progress = u.Progress
	model.Parent = int(u.Parent)
	model.Type = u.Type
	model.Lazy = u.Lazy
}
