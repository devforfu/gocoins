// Database management logic.
//
// All queries to the database are performed with the utilities from this file.
// Also, the structures representing database objects are also located here.
package server

import (
    "fmt"
    "github.com/jmoiron/sqlx"
    "github.com/lib/pq"
    _ "github.com/lib/pq"
    "sync"
    "time"
)

// A Manager is responsible for interaction with the persistent storage.
//
// The type implements application-specific data management logic. All interactions
// with the database are delegated to Manager.
type Manager struct {
    *sqlx.DB
}

func NewManager(connStr string) (*Manager, error) {
    conn, err := connect(connStr)
    if err != nil {
        return nil, err
    } else {
        return &Manager{conn}, nil
    }
}

func (db Manager) Close() error {
    return db.DB.Close()
}

// GetAvailableAccounts returns an array of all available accounts.
func (db Manager) GetAvailableAccounts() ([]Account, error) {
    var accounts []Account
    err := db.Select(&accounts, "SELECT * FROM account")
    if err != nil { return nil, err }
    return accounts, nil
}

// GetAccounts returns a subset of accounts using identifiers array to make a selection.
func (db Manager) GetAccounts(identifiers []string) ([]Account, error) {
    var accounts []Account
    err := db.Select(&accounts, "SELECT * FROM account WHERE identifier = any($1)", pq.Array(identifiers))
    if err != nil { return nil, err }
    return accounts, nil
}

// Transfer moves amount of cents between fromId and toId accounts.
func (db Manager) Transfer(fromId, toId string, amount Cents) (*Payment, error) {
    accounts, err := db.GetAccounts([]string{fromId, toId})
    if err != nil { return nil, err }

    fromAcc, toAcc := accounts[0], accounts[1]
    if fromAcc.Currency != toAcc.Currency {
        return nil, fmt.Errorf("cannot trasnfer money between accounts with different currency")
    }
    if fromAcc.Amount < amount {
        return nil, fmt.Errorf("cannot make a transaction: insufficient funds")
    }

    var mutex sync.Mutex
    mutex.Lock()
    fromAcc.Amount -= amount
    toAcc.Amount += amount
    mutex.Unlock()

    tx, err := db.Beginx()
    if err != nil { return nil, err }

    _, err = tx.NamedExec("UPDATE account SET amount = :amount WHERE identifier = :identifier", fromAcc)
    if err != nil {
        mustRollback(tx)
        return nil, err
    }

    _, err = tx.NamedExec("UPDATE account SET amount = :amount WHERE identifier = :identifier", toAcc)
    if err != nil {
        mustRollback(tx)
        return nil, err
    }

    payment := Payment{
        From:fromAcc.ID,
        To:toAcc.ID,
        Time:time.Now().UTC(),
        Amount:amount, Currency:fromAcc.Currency}

    _, err = tx.NamedExec(`
        INSERT INTO payment (from_id, to_id, transaction_time_utc, amount, currency)
        VALUES (:from_id, :to_id, :transaction_time_utc, :amount, :currency)
        `, payment)

    if err != nil {
        mustRollback(tx)
        return nil, err
    }

    if err = tx.Commit(); err != nil {
        mustRollback(tx)
        return nil, err
    }

    return &payment, nil
}

func mustRollback(tx *sqlx.Tx) {
    if err := tx.Rollback(); err != nil {
        panic(fmt.Sprintf("transaction failure: %s", err))
    }
}

// Cents stores the amount of money available in the account using integer
// data type to escape the possible rounding errors with floating point numbers.
// The whole sum is stores in number of "cents", i.e. 1/100 fraction of the money unit.
// For example, if the account sum is equal to $12.34 then it is stored as 1234 cents.
type Cents int64

type Account struct {
    ID int            `db:"user_id"`
    Identifier string `db:"identifier"`
    Currency string   `db:"currency"`
    Amount Cents      `db:"amount"`
    Created time.Time `db:"created_on"`
}

type Payment struct {
    ID int          `db:"payment_id"`
    From int        `db:"from_id"`
    To int          `db:"to_id"`
    Time time.Time  `db:"transaction_time_utc"`
    Amount Cents    `db:"amount"`
    Currency string `db:"currency"`
}

func (c Cents) String() string {
    whole, decimal := c / 100, c % 100
    return fmt.Sprintf("%s.%s", whole, decimal)
}

// connect makes a connection to the PostgreSQL database.
// The error is returned in case of any issues with the connection.
func connect(connStr string) (*sqlx.DB, error) {
    db, err := sqlx.Connect("postgres", connStr)
    if err != nil {
        return nil, fmt.Errorf("db connection error: %s", err)
    }
    return db, nil
}