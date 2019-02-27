package main

import (
    "log"
    "net/http"
)

func main() {
    mux := http.ServeMux{}
    mux.HandleFunc("/echo", echo)
    log.Fatal(http.ListenAndServe(":80", &mux))
}

func echo(w http.ResponseWriter, req *http.Request) {
    log.Printf("connected: %s", req.RemoteAddr)
    if _, err := w.Write([]byte("echo")); err != nil {
        log.Fatalf("error sending response: %s", err)
    }
}
