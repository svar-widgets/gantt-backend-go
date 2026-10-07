package main

import (
	"encoding/json"
	"fmt"
	"gantt-backend-go/data"
	"net/http"
	"time"

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
		clientID := getClientID(r)
		payload := data.TaskAddPayload{}
		err := parseForm(w, r.Body, &payload)
		if err != nil {
			sendResponse(w, nil, err)
			return
		}
		result, err := OpTaskAdd(dao, payload)
		sendResponse(w, &Response{ID: result.ID}, err)
		hub.Broadcast(clientID, result.Events)
	})

	r.Put("/tasks/{id}", func(w http.ResponseWriter, r *http.Request) {
		clientID := getClientID(r)
		id := numberParam(r, "id")
		payload := data.TaskUpdatePayload{}
		err := parseForm(w, r.Body, &payload)
		if err != nil {
			sendResponse(w, nil, err)
			return
		}
		result, err := OpTaskUpdate(dao, id, payload)
		sendResponse(w, &Response{ID: result.ID}, err)
		hub.Broadcast(clientID, result.Events)
	})

	r.Delete("/tasks/{id}", func(w http.ResponseWriter, r *http.Request) {
		clientID := getClientID(r)
		id := numberParam(r, "id")
		result, err := OpTaskDelete(dao, id)
		sendResponse(w, &Response{}, err)
		hub.Broadcast(clientID, result.Events)
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
		tids := make([]int, 0, len(tasks))
		for _, t := range tasks {
			tids = append(tids, t.ID)
		}
		data, err := dao.Links.GetByTasks(tids)
		sendResponse(w, data, err)
	})

	r.Post("/links", func(w http.ResponseWriter, r *http.Request) {
		clientID := getClientID(r)
		payload := data.LinkUpdate{}
		err := parseForm(w, r.Body, &payload)
		if err != nil {
			sendResponse(w, nil, err)
			return
		}
		result, err := OpLinkAdd(dao, payload)
		sendResponse(w, &Response{ID: result.ID}, err)
		hub.Broadcast(clientID, result.Events)
	})

	r.Put("/links/{id}", func(w http.ResponseWriter, r *http.Request) {
		clientID := getClientID(r)
		id := numberParam(r, "id")
		payload := data.LinkUpdate{}
		err := parseForm(w, r.Body, &payload)
		if err != nil {
			sendResponse(w, nil, err)
			return
		}
		result, err := OpLinkUpdate(dao, id, payload)
		sendResponse(w, &Response{ID: result.ID}, err)
		hub.Broadcast(clientID, result.Events)
	})

	r.Delete("/links/{id}", func(w http.ResponseWriter, r *http.Request) {
		clientID := getClientID(r)
		id := numberParam(r, "id")
		result, err := OpLinkDelete(dao, id)
		sendResponse(w, &Response{}, err)
		hub.Broadcast(clientID, result.Events)
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
		clientID := getClientID(r)
		payload := data.AssignmentPayload{}
		err := parseForm(w, r.Body, &payload)
		if err != nil {
			sendResponse(w, nil, err)
			return
		}
		result, err := OpAssignmentAdd(dao, payload)
		sendResponse(w, &Response{ID: result.ID}, err)
		hub.Broadcast(clientID, result.Events)
	})

	r.Put("/assignments/{id}", func(w http.ResponseWriter, r *http.Request) {
		clientID := getClientID(r)
		id := numberParam(r, "id")
		payload := data.AssignmentPayload{}
		err := parseForm(w, r.Body, &payload)
		if err != nil {
			sendResponse(w, nil, err)
			return
		}
		result, err := OpAssignmentUpdate(dao, id, payload)
		sendResponse(w, &Response{ID: result.ID}, err)
		hub.Broadcast(clientID, result.Events)
	})

	r.Delete("/assignments/{id}", func(w http.ResponseWriter, r *http.Request) {
		clientID := getClientID(r)
		id := numberParam(r, "id")
		result, err := OpAssignmentDelete(dao, id)
		sendResponse(w, &Response{}, err)
		hub.Broadcast(clientID, result.Events)
	})

	r.Post("/batch", func(w http.ResponseWriter, r *http.Request) {
		clientID := getClientID(r)
		batch := []data.BatchItem{}
		err := parseForm(w, r.Body, &batch)
		if err != nil {
			sendResponse(w, nil, err)
			return
		}
		responses, events := OpBatch(dao, batch)
		sendResponse(w, responses, nil)
		hub.Broadcast(clientID, events)
	})

	r.Get("/events", func(w http.ResponseWriter, r *http.Request) {
		clientID := r.URL.Query().Get("clientId")
		if clientID == "" {
			http.Error(w, "clientId is required", http.StatusBadRequest)
			return
		}

		stream := http.NewResponseController(w)

		sub := hub.Connect(clientID)
		defer hub.Disconnect(sub)

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, ":connected\n\n")
		if err := stream.Flush(); err != nil {
			return
		}

		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case event := <-sub.Events:
				data, err := json.Marshal(event)
				if err != nil {
					continue
				}
				fmt.Fprintf(w, "data:%s\n\n", data)
				if err := stream.Flush(); err != nil {
					return
				}
			case <-ticker.C:
				fmt.Fprint(w, ":ping\n\n")
				if err := stream.Flush(); err != nil {
					return
				}
			case <-sub.Done():
				return
			case <-r.Context().Done():
				return
			}
		}
	})
}

// The client id only decides which browser tab skips its own echo. It is
// supplied by the client, never verified, and must not be used for
// authorization or to identify a user. In production derive identity from the
// session and treat this purely as a per-tab correlation token.
func getClientID(r *http.Request) string {
	return r.Header.Get("X-Client-Id")
}
