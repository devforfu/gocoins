// Convenience tool to serialize api responses into valid JSON binary data.
package server

import (
    "encoding/json"
    "log"
    "net/http"
)

// Responder serializes API-specific data structures into binary format and
// writes them using response writer.
type Responder struct {
    http.ResponseWriter
    *json.Encoder
}

func (r Responder) SendError(message string) {
    r.WriteHeader(http.StatusBadRequest)
    err := r.Encode(Response{"error": message})
    if err != nil { log.Print(err) }
}

func (r Responder) SendSuccess(resp Response) {
    r.WriteHeader(http.StatusAccepted)
    err := r.Encode(resp)
    if err != nil { log.Print(err) }
}

// Creates a new responder to serialize the data into ResponseWriter.
func NewJSONResponse(w http.ResponseWriter) Responder {
    resp := Responder{}
    resp.ResponseWriter = w
    resp.Encoder = json.NewEncoder(w)
    return resp
}