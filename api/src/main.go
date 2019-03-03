package main

import (
    "./server"
    "context"
    "fmt"
    "log"
    "net/http"
    "os"
    "strconv"
)

func main() {
    conf := server.Config{Host:"", Port:mustGetPort(), DatabaseConn:connString()}
    srv := server.NewBillingAPI(conf)
    if err := srv.ListenAndServe(); err != http.ErrServerClosed {
        log.Fatalf("server error: %s", err)
    }
    _ = srv.Shutdown(context.TODO())
}

// connString builds a connection string using the environment variables.
func connString() string {
    var (
        host = os.Getenv("DB_HOST")
        port = os.Getenv("DB_PORT")
        dbname = os.Getenv("DB_NAME")
        user = os.Getenv("DB_USER")
        password = os.Getenv("DB_PASSWORD")
        sslmode = os.Getenv("SSL_MODE")
    )
    return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
        host, port, user, password, dbname, sslmode)
}

func mustGetPort() int {
    if port, err := strconv.Atoi(os.Getenv("PORT")); err != nil {
        panic(err)
    } else {
        return port
    }
}