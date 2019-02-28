package server

import (
    "database/sql"
    "fmt"
    "github.com/lib/pq"
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

type Cursor struct {
    *sql.DB
}

func NewCursor(connStr string) (*Cursor , error) {
    conn, err := connect(connStr)
    if err != nil {
        return nil, err
    } else {
        return &Cursor{conn}, nil
    }
}

func (db Cursor) GetAvailableAccounts() ([]Account, error) {
    rows, err := db.Query("SELECT identifier FROM account")
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    identifiers := make([]string, 0)
    for rows.Next() {
        var identifier string
        err := rows.Scan(&identifier)
        if err != nil {
            return nil, err
        }
        identifiers = append(identifiers, identifier)
    }
    return db.GetAccounts(identifiers)
}

func (db Cursor) GetAccounts(identifiers []string) ([]Account, error) {
    rows, err := db.Query("SELECT * FROM account WHERE identifier = ANY ($1)", pq.Array(identifiers))
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    accounts := make([]Account, 0)
    for rows.Next() {
        var account Account
        err := rows.Scan(&account.ID, &account.Identifier, &account.Currency, &account.Created)
        if err != nil {
            return nil, err
        }
        accounts = append(accounts, account)
    }
    return accounts, nil
}

func (db Cursor) GetSingleAccount(identifier string) (acc *Account, err error) {
    row := db.QueryRow("SELECT * FROM account WHERE identifier = $1", identifier)
    if err := row.Scan(acc); err != nil {
        return nil, err
    }
    return acc, nil
}

func (db Cursor) Transfer(fromId, toId string, amount cents) (string, error) {
    fromAcc, err := db.GetSingleAccount(fromId)
    if err != nil { return "", err }

    toAcc, err := db.GetSingleAccount(toId)
    if err != nil { return "", err }

    if fromAcc.Currency != toAcc.Currency {
        return "", fmt.Errorf("accounts have different currency")
    }

    if fromAcc.Amount < amount {
        return "", fmt.Errorf("not enough funds")
    }

    return "", fmt.Errorf("not implemented")
}

type cents int64

type Account struct {
    ID int
    Identifier string
    Currency string
    Amount cents
    Created time.Time
}

type Payment struct {
    ID int
    From, To Account
    TransactionTime time.Time
    Amount cents
    Currency string
}

func (c cents) String() string {
    whole, decimal := c / 100, c % 100
    return fmt.Sprintf("%s.%s", whole, decimal)
}
