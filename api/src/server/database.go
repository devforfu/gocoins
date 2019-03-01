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
//
// Accounts fromId and toId should be in the same currency. Also, the account fromId
// should have sufficient amount of funds to perform a transaction. In case if any
// of these preconditions is violated, or accounts with these IDs are not found,
// then the error is returned.
//
// The process of accounts updating performed as a single transaction. In case if
// the transaction cannot be rolled back, the method panics.
func (db Manager) Transfer(fromId, toId string, amount Cents) (*Payment, error) {
    accounts, err := db.GetAccounts([]string{fromId, toId})
    if err != nil {
        return nil, inputError("cannot find the accounts")
    }

    fromAcc, toAcc := accounts[0], accounts[1]
    if fromAcc.Currency != toAcc.Currency {
        return nil, inputError("cannot transfer money between accounts with different currency")
    }
    if fromAcc.Amount < amount {
        return nil, inputError("cannot make a transaction: insufficient funds")
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
        return nil, internalError(err)
    }

    _, err = tx.NamedExec("UPDATE account SET amount = :amount WHERE identifier = :identifier", toAcc)
    if err != nil {
        mustRollback(tx)
        return nil, internalError(err)
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
        return nil, internalError(err)
    }

    if err = tx.Commit(); err != nil {
        mustRollback(tx)
        return nil, internalError(err)
    }

    return &payment, nil
}

// mustRollback panics if a transaction cannot be rolled back.
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

// Account represents information about payment system's account.
type Account struct {
    ID int            `db:"user_id"`
    Identifier string `db:"identifier"`
    Currency string   `db:"currency"`
    Amount Cents      `db:"amount"`
    Created time.Time `db:"created_on"`
}

// Payment contains an information about a money transfer between accounts.
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

// managerError is returned whenever Manager cannot successfully perform operation
// due to wrong input or some internal bug.
// Custom type helps to distinguish between these two types of errors and send
// error message to the client only in case when the error is not internal one.
type managerError struct {
    message string
    internal bool
}

func inputError(message string) managerError {
    return managerError{message, false}
}

func internalError(err error) managerError {
    return managerError{err.Error(), true}
}

func (m managerError) Error() string {
    return m.message
}
