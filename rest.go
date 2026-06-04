package main

import (
	"bytes"
	"errors"
	"fmt"
	"gantt-backend-go/data"
	"net/http"

	"github.com/go-chi/chi"
)

func initRoutes(r chi.Router, dao *data.DAO) {

	r.Get("/tasks", func(w http.ResponseWriter, r *http.Request) {
		data, err := dao.Tasks.GetBranch(0)
		sendResponse(w, data, err)
	})

	r.Get("/tasks/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := numberParam(r, "id")
		data, err := dao.Tasks.GetBranch(id)
		sendResponse(w, data, err)
	})

	r.Post("/tasks", func(w http.ResponseWriter, r *http.Request) {
		task := data.TaskAddPayload{}
		err := parseForm(w, r.Body, &task)
		if err != nil {
			sendResponse(w, nil, err)
			return
		}
		id, err := dao.Tasks.Add(task)
		if task.Mode != "" {
			err = dao.Tasks.Move(id, data.TaskUpdatePayload{
				TaskUpdate: task.Task,
				Mode:       task.Mode,
				Target:     task.Target,
			})
			if err != nil {
				sendResponse(w, nil, err)
				return
			}
		}
		sendResponse(w, &Response{id}, err)
	})

	r.Put("/tasks/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := numberParam(r, "id")
		u := data.TaskUpdatePayload{}
		err := parseForm(w, r.Body, &u)
		if err != nil {
			sendResponse(w, nil, err)
			return
		}

		if u.Operation == "copy" {
			nids, err := copyTask(dao, id, u)
			if err != nil {
				sendResponse(w, nil, err)
				return
			}
			sendResponse(w, &Response{ID: nids[0]}, err)
			return
		} else if u.Operation == "move" {
			err = dao.Tasks.Move(id, u)
			if err != nil {
				sendResponse(w, nil, err)
				return
			}
		} else {
			err = dao.Tasks.Update(id, u)
			if err != nil {
				sendResponse(w, nil, err)
				return
			}
		}

		sendResponse(w, &Response{id}, err)
	})

	r.Delete("/tasks/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := numberParam(r, "id")
		err := deleteTask(dao, id)
		sendResponse(w, &Response{}, err)
	})

	r.Get("/links", func(w http.ResponseWriter, r *http.Request) {
		tasks, err := dao.Tasks.GetBranch(0)
		if err != nil {
			sendResponse(w, nil, err)
			return
		}
		tids := make([]int, 0)
		for _, t := range tasks {
			tids = append(tids, t.ID)
		}
		data, err := dao.Links.GetBranch(tids)
		if err != nil {
			sendResponse(w, nil, err)
			return
		}
		sendResponse(w, data, err)
	})

	r.Get("/links/{taskId}", func(w http.ResponseWriter, r *http.Request) {
		id := numberParam(r, "taskId")
		tasks, err := dao.Tasks.GetBranch(id)
		if err != nil {
			sendResponse(w, nil, err)
			return
		}
		tids := make([]int, 0)
		tids = append(tids, id)
		for _, t := range tasks {
			tids = append(tids, t.ID)
		}
		data, err := dao.Links.GetBranch(tids)
		sendResponse(w, data, err)
	})

	r.Post("/links", func(w http.ResponseWriter, r *http.Request) {
		data := data.LinkUpdate{}
		err := parseForm(w, r.Body, &data)
		if err != nil {
			sendResponse(w, nil, err)
			return
		}
		id, err := dao.Links.Add(data)
		sendResponse(w, &Response{id}, err)
	})

	r.Put("/links/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := numberParam(r, "id")
		data := data.LinkUpdate{}
		err := parseForm(w, r.Body, &data)
		if err != nil {
			sendResponse(w, nil, err)
			return
		}
		err = dao.Links.Update(id, data)
		sendResponse(w, &Response{id}, err)
	})

	r.Delete("/links/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := numberParam(r, "id")
		err := dao.Links.Delete(id)
		sendResponse(w, &Response{}, err)
	})

	r.Get("/resources", func(w http.ResponseWriter, r *http.Request) {
		data, err := dao.Resources.GetAll()
		sendResponse(w, data, err)
	})

	r.Get("/assignments", func(w http.ResponseWriter, r *http.Request) {
		tasks, err := dao.Tasks.GetBranch(0)
		if err != nil {
			sendResponse(w, nil, err)
			return
		}
		tids := make([]int, 0)
		for _, t := range tasks {
			tids = append(tids, t.ID)
		}
		data, err := dao.Assignments.GetByTasks(tids)
		sendResponse(w, data, err)
	})

	r.Get("/assignments/{taskId}", func(w http.ResponseWriter, r *http.Request) {
		id := numberParam(r, "taskId")
		tasks, err := dao.Tasks.GetBranch(id)
		if err != nil {
			sendResponse(w, nil, err)
			return
		}
		tids := make([]int, 0)
		for _, t := range tasks {
			tids = append(tids, t.ID)
		}
		data, err := dao.Assignments.GetByTasks(tids)
		sendResponse(w, data, err)
	})

	r.Post("/assignments", func(w http.ResponseWriter, r *http.Request) {
		data := data.AssignmentPayload{}
		err := parseForm(w, r.Body, &data)
		if err != nil {
			sendResponse(w, nil, err)
			return
		}

		id, err := dao.Assignments.Add(data)
		sendResponse(w, &Response{id}, err)
	})

	r.Put("/assignments/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := numberParam(r, "id")
		data := data.AssignmentPayload{}
		err := parseForm(w, r.Body, &data)
		if err != nil {
			sendResponse(w, nil, err)
			return
		}
		err = dao.Assignments.Update(id, data)
		sendResponse(w, &Response{id}, err)
	})

	r.Delete("/assignments/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := numberParam(r, "id")
		err := dao.Assignments.Delete(id)
		sendResponse(w, &Response{}, err)
	})

	r.Post("/batch", func(w http.ResponseWriter, r *http.Request) {
		batch := []data.BatchItem{}
		err := parseForm(w, r.Body, &batch)
		if err != nil {
			sendResponse(w, nil, err)
			return
		}

		responses := make([]Response, 0, len(batch))
		for _, d := range batch {
			if d.Data == nil {
				sendResponse(w, nil, errors.New("data is null"))
				return
			}

			data, err := d.Data.MarshalJSON()
			if err != nil {
				sendResponse(w, nil, err)
				return
			}

			req, err := http.NewRequest(d.Method, fmt.Sprintf("%s/%s", Config.Server.URL, d.Url), bytes.NewBuffer(data))
			if err != nil {
				sendResponse(w, nil, err)
				return
			}

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				sendResponse(w, nil, err)
				return
			}
			defer resp.Body.Close()

			var response Response
			if err = parseForm(w, resp.Body, &response); err != nil {
				sendResponse(w, nil, err)
				return
			}

			responses = append(responses, response)
		}

		sendResponse(w, responses, err)
	})

}

func copyTask(dao *data.DAO, id int, u data.TaskUpdatePayload) ([]int, error) {
	ids, nids, err := dao.Tasks.Copy(id, u)
	if err != nil {
		return nil, err
	}
	if u.Lazy {
		err = dao.Links.CopyBranch(ids[1:], nids[1:])
		if err != nil {
			return nil, err
		}
	}

	return nids, nil
}

func deleteTask(dao *data.DAO, id int) error {
	removed, err := dao.Tasks.Delete(id)
	if err == nil {
		err = dao.Links.DeleteBranch(removed)
	}
	if err == nil {
		err = dao.Assignments.DeleteByTasks(removed)
	}
	return err
}
