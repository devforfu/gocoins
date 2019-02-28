package server

import (
    "fmt"
    "log"
    "net/http"
)

type Response map[string]interface{}

type Config struct {
    Host string
    Port int
    DatabaseConn string
}

func (c Config) Addr() string {
    return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

func (c Config) URL() string {
    return fmt.Sprintf("http://%s", c.Addr())
}

func Create(conf Config) *http.Server {
    mux := http.NewServeMux()
    mux.Handle("/", http.HandlerFunc(notFound))
    mux.Handle("/status", http.HandlerFunc(status))
    return &http.Server{Addr:conf.Addr(), Handler:mux}
}

func status(w http.ResponseWriter, req *http.Request) {
    log.Printf("connected: %s", req.RemoteAddr)
    resp := NewJSONResponse(w)
    resp.SendSuccess(Response{"success": true})
}

//func accounts(w http.ResponseWriter, req *http.Request) {
//
//}

func notFound(w http.ResponseWriter, req *http.Request) {
    NewJSONResponse(w).SendError("not found")
}
