package main

import (
	"gantt-backend-go/data"
	"net/http"

	"github.com/go-chi/chi"
)

func initRoutes(r chi.Router, dao *data.DAO) {

	r.Get("/tasks", func(w http.ResponseWriter, r *http.Request) {
		// data, err := dao.Tasks.GetAll()
		data, err := dao.Tasks.GetBranch(0)
		sendResponse(w, data, err)
	})

	r.Get("/tasks/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := numberParam(r, "id")
		data, err := dao.Tasks.GetBranch(id)
		sendResponse(w, data, err)
	})

	r.Post("/tasks", func(w http.ResponseWriter, r *http.Request) {
		data := data.TaskUpdate{}
		err := parseForm(w, r, &data)
		if err != nil {
			sendResponse(w, nil, err)
			return
		}
		id, err := dao.Tasks.Add(data)
		sendResponse(w, &Response{id}, err)
	})

	r.Put("/tasks/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := numberParam(r, "id")

		u := data.UpdatePayload{}
		err := parseForm(w, r, &u)
		if err != nil {
			sendResponse(w, nil, err)
			return
		}

		if u.Operation == "copy" {
			ids, nids, err := dao.Tasks.Copy(id, u)
			if err != nil {
				sendResponse(w, nil, err)
				return
			}
			if u.Nested {
				err = dao.Links.CopyBranch(ids[1:], nids[1:])
				if err != nil {
					sendResponse(w, nil, err)
					return
				}
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

		removed, err := dao.Tasks.Delete(id)
		if err == nil {
			err = dao.Links.DeleteBranch(removed)
		}

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
		err := parseForm(w, r, &data)
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
		err := parseForm(w, r, &data)
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

}
