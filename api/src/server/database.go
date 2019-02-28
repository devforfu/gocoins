package server

import (
    "database/sql"
    "fmt"
    _ "github.com/lib/pq"
    "time"
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

type decimal int64

type Account struct {
    ID int
    Identifier string
    Currency string
    Created time.Time
}

type Payment struct {
    ID int
    From, To Account
    TransactionTime time.Time
    Amount decimal
    Currency string
}

func (d decimal) String() string {
    whole, decimal := d / 100, d % 100
    return fmt.Sprintf("%s.%s", whole, decimal)
}
