package main

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
)

type Response struct {
	ID int `json:"id,omitempty"`
}

func numberParam(r *http.Request, key string) int {
	value := chi.URLParam(r, key)
	num, _ := strconv.Atoi(value)

	return num
}

func parseForm(w http.ResponseWriter, r *http.Request, o interface{}) error {
	body := http.MaxBytesReader(w, r.Body, 1048576)
	dec := json.NewDecoder(body)
	err := dec.Decode(&o)

	return err
}

func sendResponse(w http.ResponseWriter, data interface{}, err error) {
	if err != nil {
		format.Text(w, 500, err.Error())
	} else {
		format.JSON(w, 200, data)
	}
}
