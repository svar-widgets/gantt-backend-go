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
	Text                  string          `json:"text"`
	Details               string          `json:"details"`
	Start                 *common.JDate   `json:"start"`
	End                   *common.JDate   `json:"end"`
	Duration              int             `json:"duration"`
	Progress              int             `json:"progress"`
	Parent                common.FuzzyInt `json:"parent"`
	Type                  string          `json:"type"`
	Lazy                  bool            `json:"lazy"`
	Constraint            *Constraint     `json:"constraint"`
	Deadline              *common.JDate   `json:"deadline"`
	BaseStart             *common.JDate   `json:"base_start"`
	BaseEnd               *common.JDate   `json:"base_end"`
	BaseDuration          int             `json:"base_duration"`
	Calendar              *common.ID      `json:"calendar"`
	SkipResourceCalendars bool            `json:"skipResourceCalendars"`
	Unscheduled           bool            `json:"unscheduled"`
	Manual                bool            `json:"manual"`
	Inactive              bool            `json:"inactive"`
	Rollup                bool            `json:"rollup"`
	Segments              Segments        `json:"segments"`
}

type TaskUpdatePayload struct {
	TaskUpdate

	Operation string `json:"operation"`
	Target    int    `json:"target"`
	Mode      string `json:"mode"`
	Index     *int   `json:"index"`
}

type TaskAddPayload struct {
	Task TaskUpdate `json:"task"`

	Target int    `json:"target"`
	Mode   string `json:"mode"`
	Index  *int   `json:"index"`
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

func (d *TasksDAO) Add(data TaskAddPayload) (int, error) {
	task := Task{}
	data.Task.fillModel(&task)
	err := d.insert(&task, data.Index)
	return task.ID, err
}

// places the task at the given index, or last in its branch
func (d *TasksDAO) insert(task *Task, index *int) error {
	if index != nil {
		task.Index = *index
		return d.db.Create(task).Error
	}

	return d.db.Transaction(func(tx *gorm.DB) error {
		branch, err := d.getKids(tx, task.Parent)
		if err != nil {
			return err
		}
		task.Index = len(branch)
		return tx.Create(task).Error
	})
}

func (d *TasksDAO) Update(id int, data TaskUpdatePayload) error {
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

func (d *TasksDAO) Move(id int, data TaskUpdatePayload) error {
	task, err := d.GetOne(id)
	if err != nil {
		return err
	}
	target, err := d.GetOne(data.Target)
	if err != nil {
		return err
	}

	return d.db.Transaction(func(tx *gorm.DB) error {
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
				data.Mode = "after"
				target = targetBranch[l-1]
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
						ind++
						branchUpd[ind] = task
						ind++
						continue
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
	})
}

func (d *TasksDAO) Copy(id int, data TaskUpdatePayload) ([]int, []int, error) {
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

	if data.Index != nil {
		return d.createCopy(nil, task, targetParent, data.Lazy, data.Index)
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

	ids, nids, err := d.createCopy(nil, task, targetParent, data.Lazy, nil)
	if err != nil {
		return nil, nil, err
	}

	ntask, err := d.GetOne(nids[0])
	if err != nil {
		return nil, nil, err
	}

	err = d.db.Transaction(func(tx *gorm.DB) error {
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

		return d.refreshBranchOrder(tx, branchUpd)
	})
	if err != nil {
		return nil, nil, err
	}

	return ids, nids, nil
}

func (d *TasksDAO) createCopy(tx *gorm.DB, task Task, parent int, lazy bool, index *int) ([]int, []int, error) {
	ids := make([]int, 0)
	nids := make([]int, 0)
	// a model, not a payload: lazy is branch state, which fillModel never takes from a client
	ntask := Task{
		Text:                  task.Text,
		Details:               task.Details,
		Start:                 task.Start,
		End:                   task.End,
		Duration:              task.Duration,
		Progress:              task.Progress,
		Parent:                parent,
		Type:                  task.Type,
		Lazy:                  task.Lazy,
		Constraint:            task.Constraint,
		Deadline:              task.Deadline,
		BaseStart:             task.BaseStart,
		BaseEnd:               task.BaseEnd,
		BaseDuration:          task.BaseDuration,
		Calendar:              task.Calendar,
		SkipResourceCalendars: task.SkipResourceCalendars,
		Unscheduled:           task.Unscheduled,
		Manual:                task.Manual,
		Inactive:              task.Inactive,
		Rollup:                task.Rollup,
		Segments:              task.Segments,
	}
	if err := d.insert(&ntask, index); err != nil {
		return nil, nil, err
	}
	nid := ntask.ID
	ids = append(ids, task.ID)
	nids = append(nids, nid)

	if lazy {
		kids, err := d.getKids(tx, task.ID)
		if err != nil {
			return nil, nil, err
		}
		for i, kid := range kids {
			kidIndex := i
			oids, knids, err := d.createCopy(tx, kid, nid, lazy, &kidIndex)
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
	model.Details = u.Details
	model.Start = u.Start
	model.End = u.End
	model.Duration = u.Duration
	model.Progress = u.Progress
	model.Parent = int(u.Parent)
	model.Type = u.Type
	model.Constraint = u.Constraint
	model.Deadline = u.Deadline
	model.BaseStart = u.BaseStart
	model.BaseEnd = u.BaseEnd
	model.BaseDuration = u.BaseDuration
	model.Calendar = u.Calendar
	model.SkipResourceCalendars = u.SkipResourceCalendars
	model.Unscheduled = u.Unscheduled
	model.Manual = u.Manual
	model.Inactive = u.Inactive
	model.Rollup = u.Rollup
	model.Segments = u.Segments
}
