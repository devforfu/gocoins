package main

import (
    "./server"
    "context"
    "database/sql"
    "fmt"
    _ "github.com/lib/pq"
    "log"
    "net/http"
)

const (
    host = "db"
    port = 5432
    user = "docker"
    password = "docker"
    dbname = "docker"
    sslmode = "disable"
)

func connString() string {
    return fmt.Sprintf(
        "host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
        host, port, user, password, dbname, sslmode)
}

func main() {
    conf := server.Config{Host:"", Port:80, DatabaseConn:connString()}
    srv := server.Create(conf)
    if err := srv.ListenAndServe(); err != http.ErrServerClosed {
        log.Fatalf("server error: %s", err)
    }
    _ = srv.Shutdown(context.TODO())
}

func connect() (*sql.DB, error) {
    db, err := sql.Open("postgres", connString())
    if err != nil {
        return nil, fmt.Errorf("conn string error: %s", err)
    }
    if err := db.Ping(); err != nil {
        return nil, fmt.Errorf("db ping error: %s", err)
    }
    return db, nil
}
