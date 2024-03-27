package data

import (
	"fmt"
	"gantt-backend-go/common"

	"gorm.io/gorm"
)

type LinksDAO struct {
	db *gorm.DB
}

type LinkUpdate struct {
	Source common.FuzzyInt `json:"source"`
	Target common.FuzzyInt `json:"target"`
	Type   string          `json:"type"`
}

func NewLinksDAO(db *gorm.DB) *LinksDAO {
	return &LinksDAO{db}
}

func (d *LinksDAO) GetOne(id int) (Link, error) {
	link := Link{}
	err := d.db.Find(&link, id).Error
	if link.ID == 0 {
		return Link{}, fmt.Errorf("link with id %d not found", id)
	}

	return link, err
}

func (d *LinksDAO) GetAll() ([]Link, error) {
	links := make([]Link, 0)
	err := d.db.Find(&links).Error

	return links, err
}

func (d *LinksDAO) GetBranch(tasks []int) ([]Link, error) {
	links := make([]Link, 0)
	err := d.db.Where("source IN ? AND target IN ?", tasks, tasks).Find(&links).Error
	if err != nil {
		return nil, err
	}
	return links, err
}

func (d *LinksDAO) Add(data LinkUpdate) (int, error) {
	link := Link{}
	data.fillModel(&link)
	err := d.db.Create(&link).Error

	return link.ID, err
}

func (d *LinksDAO) Update(id int, data LinkUpdate) error {
	link, err := d.GetOne(id)
	if err != nil {
		return err
	}
	data.fillModel(&link)
	err = d.db.Save(&link).Error

	return err
}

func (d *LinksDAO) Delete(id int) error {
	err := d.db.Delete(&Link{}, id).Error
	return err
}

func (d *LinksDAO) DeleteBranch(tasks []int) error {
	err := d.db.Where("source IN ? OR target IN ?", tasks, tasks).Delete(&Link{}).Error
	return err
}

func (d *LinksDAO) CopyBranch(old []int, new []int) error {
	links, err := d.GetAll()
	if err == nil {
		sources := make([]Link, 0)
		targets := make([]Link, 0)
		for _, link := range links {
			for ind, o := range old {
				if link.Source == o {
					link.Source = new[ind]
					sources = append(sources, link)
				} else if link.Target == o {
					link.Target = new[ind]
					targets = append(targets, link)
				}
			}
		}

		targetsLen := len(targets)
		for i, source := range sources {
			t := common.Search(targetsLen, func(index int) bool {
				return targets[index].ID == source.ID
			})
			if t > -1 {
				sources[i].Target = targets[t].Target
				targets = append(targets[:t], targets[t+1:]...)
			}
		}

		toCopy := make([]Link, 0)
		toCopy = append(toCopy, sources...)
		toCopy = append(toCopy, targets...)

		for _, link := range toCopy {
			_, err = d.Add(LinkUpdate{
				Source: common.FuzzyInt(link.Source),
				Target: common.FuzzyInt(link.Target),
				Type:   link.Type,
			})
			if err != nil {
				break
			}
		}
	}
	return err
}

func (u *LinkUpdate) fillModel(model *Link) {
	model.Source = int(u.Source)
	model.Target = int(u.Target)
	model.Type = u.Type
}
