package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"gantt-backend-go/data"
	"strconv"
	"strings"
)

type OpResult struct {
	ID     int
	Events []Event
}

func (result *OpResult) AddEvent(event Event) {
	result.Events = append(result.Events, event)
}

// ---------- Tasks ----------

func OpTaskAdd(dao *data.DAO, payload data.TaskAddPayload) (OpResult, error) {
	var result OpResult
	err := dao.Transaction(func(dao *data.DAO) error {
		if payload.Task.Parent != 0 {
			if _, err := dao.Tasks.GetOne(int(payload.Task.Parent)); err != nil {
				return err
			}
		}
		if payload.Target != 0 && payload.Mode != "" {
			if _, err := dao.Tasks.GetOne(payload.Target); err != nil {
				return err
			}
		}
		id, err := dao.Tasks.Add(payload)
		if err != nil {
			return err
		}
		if payload.Target != 0 && payload.Mode != "" {
			err = dao.Tasks.Move(id, data.TaskUpdatePayload{
				TaskUpdate: payload.Task,
				Mode:       payload.Mode,
				Target:     payload.Target,
			})
			if err != nil {
				return err
			}
		}
		result = OpResult{ID: id}
		if task, err := dao.Tasks.GetOne(id); err == nil {
			index := task.Index
			result.AddEvent(Event{
				Type:   "task",
				Action: "add",
				ID:     id,
				Data:   TaskAddEvent{Task: task, Target: payload.Target, Mode: payload.Mode, Index: &index},
			})
		}
		return nil
	})
	if err != nil {
		return OpResult{}, err
	}
	return result, nil
}

func OpTaskUpdate(dao *data.DAO, id int, payload data.TaskUpdatePayload) (OpResult, error) {
	switch payload.Operation {
	case "copy":
		return opTaskCopy(dao, id, payload)
	case "move":
		return opTaskMove(dao, id, payload)
	default:
		return opTaskUpdate(dao, id, payload)
	}
}

func opTaskCopy(dao *data.DAO, id int, payload data.TaskUpdatePayload) (OpResult, error) {
	var result OpResult
	err := dao.Transaction(func(dao *data.DAO) error {
		ids, nids, err := dao.Tasks.Copy(id, payload)
		if err != nil {
			return err
		}
		newLinkIDs := []int{}
		if payload.Lazy {
			newLinkIDs, err = dao.Links.CopyBranch(ids[1:], nids[1:])
			if err != nil {
				return err
			}
		}
		result = OpResult{ID: nids[0]}
		for i, nid := range nids {
			newTask, err := dao.Tasks.GetOne(nid)
			if err != nil {
				continue
			}
			index := newTask.Index
			event := TaskAddEvent{Task: newTask, Index: &index}
			if i == 0 {
				event.Target = payload.Target
				event.Mode = payload.Mode
				if payload.Lazy {
					event.Lazy = false
				}
			}
			result.AddEvent(Event{Type: "task", Action: "add", ID: nid, Data: event})
		}
		for _, lid := range newLinkIDs {
			if link, err := dao.Links.GetOne(lid); err == nil {
				result.AddEvent(Event{Type: "link", Action: "add", ID: lid, Data: link})
			}
		}
		return nil
	})
	if err != nil {
		return OpResult{}, err
	}
	return result, nil
}

func opTaskMove(dao *data.DAO, id int, payload data.TaskUpdatePayload) (OpResult, error) {
	err := dao.Tasks.Move(id, payload)
	if err != nil {
		return OpResult{}, err
	}
	result := OpResult{ID: id}
	if newTask, tErr := dao.Tasks.GetOne(id); tErr == nil {
		event := TaskUpdateEvent{Task: newTask, Operation: "move", Target: payload.Target, Mode: payload.Mode}
		if kids, kErr := dao.Tasks.GetBranch(id); kErr == nil && len(kids) > 0 {
			event.Lazy = true
		}
		result.AddEvent(Event{Type: "task", Action: "update", ID: id, Data: event})
	}
	return result, nil
}

func opTaskUpdate(dao *data.DAO, id int, payload data.TaskUpdatePayload) (OpResult, error) {
	err := dao.Tasks.Update(id, payload)
	if err != nil {
		return OpResult{}, err
	}
	result := OpResult{ID: id}
	if newTask, tErr := dao.Tasks.GetOne(id); tErr == nil {
		event := TaskUpdateEvent{Task: newTask, Operation: ""}
		result.AddEvent(Event{Type: "task", Action: "update", ID: id, Data: event})
	}
	return result, nil
}

func OpTaskDelete(dao *data.DAO, id int) (OpResult, error) {
	err := dao.Transaction(func(dao *data.DAO) error {
		removed, err := dao.Tasks.Delete(id)
		if err == nil {
			err = dao.Links.DeleteBranch(removed)
		}
		if err == nil {
			err = dao.Assignments.DeleteByTasks(removed)
		}
		return err
	})
	if err != nil {
		return OpResult{}, err
	}
	result := OpResult{}
	result.AddEvent(Event{Type: "task", Action: "delete", ID: id})
	return result, nil
}

// ---------- Links ----------

func OpLinkAdd(dao *data.DAO, payload data.LinkUpdate) (OpResult, error) {
	id, err := dao.Links.Add(payload)
	if err != nil {
		return OpResult{}, err
	}
	result := OpResult{ID: id}
	if link, err := dao.Links.GetOne(id); err == nil {
		result.AddEvent(Event{Type: "link", Action: "add", ID: id, Data: link})
	}
	return result, nil
}

func OpLinkUpdate(dao *data.DAO, id int, payload data.LinkUpdate) (OpResult, error) {
	err := dao.Links.Update(id, payload)
	if err != nil {
		return OpResult{}, err
	}
	result := OpResult{ID: id}
	if link, err := dao.Links.GetOne(id); err == nil {
		result.AddEvent(Event{Type: "link", Action: "update", ID: id, Data: link})
	}
	return result, nil
}

func OpLinkDelete(dao *data.DAO, id int) (OpResult, error) {
	err := dao.Links.Delete(id)
	if err != nil {
		return OpResult{}, err
	}
	result := OpResult{}
	result.AddEvent(Event{Type: "link", Action: "delete", ID: id})
	return result, nil
}

// ---------- Assignments ----------

func OpAssignmentAdd(dao *data.DAO, payload data.AssignmentPayload) (OpResult, error) {
	id, err := dao.Assignments.Add(payload)
	if err != nil {
		return OpResult{}, err
	}
	result := OpResult{ID: id}
	if assignment, err := dao.Assignments.GetOne(id); err == nil {
		result.AddEvent(Event{Type: "assignment", Action: "add", ID: id, Data: assignment})
	}
	return result, nil
}

func OpAssignmentUpdate(dao *data.DAO, id int, payload data.AssignmentPayload) (OpResult, error) {
	err := dao.Assignments.Update(id, payload)
	if err != nil {
		return OpResult{}, err
	}
	result := OpResult{ID: id}
	if assignment, err := dao.Assignments.GetOne(id); err == nil {
		result.AddEvent(Event{Type: "assignment", Action: "update", ID: id, Data: assignment})
	}
	return result, nil
}

func OpAssignmentDelete(dao *data.DAO, id int) (OpResult, error) {
	err := dao.Assignments.Delete(id)
	if err != nil {
		return OpResult{}, err
	}
	result := OpResult{}
	result.AddEvent(Event{Type: "assignment", Action: "delete", ID: id})
	return result, nil
}

// ---------- Batching ----------

type batchOpFn func(dao *data.DAO, id int, raw json.RawMessage) (OpResult, error)

func OpBatch(dao *data.DAO, items []data.BatchItem) ([]BatchResponse, []Event) {
	responses := make([]BatchResponse, len(items))
	events := []Event{}
	failed := -1

	err := dao.Transaction(func(dao *data.DAO) error {
		for i, item := range items {
			result, err := opBatchItem(dao, item)
			if err != nil {
				failed = i
				return err
			}
			responses[i] = BatchResponse{ID: result.ID}
			events = append(events, result.Events...)
		}
		return nil
	})
	if err == nil {
		return responses, events
	}

	if failed < 0 {
		failed = 0
	}
	for i := range responses {
		switch {
		case i == failed:
			responses[i] = BatchResponse{Error: err.Error()}
		case i < failed:
			responses[i] = BatchResponse{Error: fmt.Sprintf("rolled back: request %d failed", failed)}
		default:
			responses[i] = BatchResponse{Error: fmt.Sprintf("not attempted: request %d failed", failed)}
		}
	}

	return responses, nil
}

func opBatchItem(dao *data.DAO, item data.BatchItem) (OpResult, error) {
	entity, id, err := parseBatchUrl(item.Url)
	if err != nil {
		return OpResult{}, fmt.Errorf("%s %s: %w", item.Method, item.Url, err)
	}
	op, ok := batchOps[item.Method+" /"+entity]
	if !ok {
		return OpResult{}, fmt.Errorf("unsupported request %s %s", item.Method, item.Url)
	}
	result, err := op(dao, id, item.Data)
	if err != nil {
		return OpResult{}, fmt.Errorf("%s %s: %w", item.Method, item.Url, err)
	}
	return result, nil

}

func parseBatchUrl(url string) (string, int, error) {
	parts := strings.Split(strings.Trim(url, "/"), "/")
	switch len(parts) {
	case 1:
		return parts[0], 0, nil
	case 2:
		id, err := strconv.Atoi(parts[1])
		if err != nil || id <= 0 {
			return "", 0, fmt.Errorf("invalid id %q", parts[1])
		}
		return parts[0], id, nil
	default:
		return "", 0, fmt.Errorf("invalid url %q", url)
	}
}

func addOp[Payload any](op func(*data.DAO, Payload) (OpResult, error)) batchOpFn {
	return func(dao *data.DAO, id int, raw json.RawMessage) (OpResult, error) {
		if id != 0 {
			return OpResult{}, errors.New("id not allowed")
		}
		var payload Payload
		if err := json.Unmarshal(raw, &payload); err != nil {
			return OpResult{}, err
		}
		return op(dao, payload)
	}
}

func updateOp[Payload any](op func(*data.DAO, int, Payload) (OpResult, error)) batchOpFn {
	return func(dao *data.DAO, id int, raw json.RawMessage) (OpResult, error) {
		if id == 0 {
			return OpResult{}, errors.New("id required")
		}
		var payload Payload
		if err := json.Unmarshal(raw, &payload); err != nil {
			return OpResult{}, err
		}
		return op(dao, id, payload)
	}
}

func deleteOp(op func(*data.DAO, int) (OpResult, error)) batchOpFn {
	return func(dao *data.DAO, id int, _ json.RawMessage) (OpResult, error) {
		if id == 0 {
			return OpResult{}, errors.New("id required")
		}
		return op(dao, id)
	}
}

var batchOps = map[string]batchOpFn{
	"POST /tasks":         addOp(OpTaskAdd),
	"PUT /tasks":          updateOp(OpTaskUpdate),
	"DELETE /tasks":       deleteOp(OpTaskDelete),
	"POST /links":         addOp(OpLinkAdd),
	"PUT /links":          updateOp(OpLinkUpdate),
	"DELETE /links":       deleteOp(OpLinkDelete),
	"POST /assignments":   addOp(OpAssignmentAdd),
	"PUT /assignments":    updateOp(OpAssignmentUpdate),
	"DELETE /assignments": deleteOp(OpAssignmentDelete),
}
