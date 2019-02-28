package server

import (
    "database/sql"
    "fmt"
    _ "github.com/lib/pq"
)

func connect(connStr string) (*sql.DB, error) {
    db, err := sql.Open("postgres", connStr)
    if err != nil {
        return nil, fmt.Errorf("conn string error: %s", err)
    }
    if err := db.Ping(); err != nil {
        return nil, fmt.Errorf("db ping error: %s", err)
    }
    return db, nil
}
