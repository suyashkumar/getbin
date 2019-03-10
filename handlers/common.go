package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

func sendErrorWithCode(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "text/plain")
	w.Header().Add("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", "error.txt"))
	w.WriteHeader(code)
	_, err := fmt.Fprintln(w, msg)
	if err != nil {
		log.Println("ERROR: error writing to response writer in sendError")
	}
}

func sendJSON(w http.ResponseWriter, v interface{}, statusCode int) error {
	resBytes, err := json.Marshal(v)
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Add("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", "error.json"))
	w.WriteHeader(statusCode)
	fmt.Fprintf(w, string(resBytes))
	return nil
}