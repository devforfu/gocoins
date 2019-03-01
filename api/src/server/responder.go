// Convenience tool to serialize the API responses into a valid JSON binary data.
package server

import (
    "encoding/json"
    "fmt"
    "log"
    "net/http"
)

// Responder serializes API-specific data structures into binary format and
// writes them using response writer.
type Responder struct {
    http.ResponseWriter
    *json.Encoder
}

func (r Responder) SendRequestError(message string) {
    r.SendError(fmt.Errorf(message), http.StatusBadRequest)
}

func (r Responder) SendServerError(message string) {
    r.SendError(fmt.Errorf(message), http.StatusInternalServerError)
}

func (r Responder) SendError(err error, status int) {
    r.WriteHeader(status)
    err = r.Encode(Response{"error": err.Error(), "status": status})
    if err != nil { log.Printf("encoding error: %s", err) }
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