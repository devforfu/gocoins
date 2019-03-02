package server

import (
    "fmt"
    "testing"
    "time"
)

func TestGettingListOfAccounts(t *testing.T) {

}


// ---------------------
// Manager class mockery
// ---------------------


func init() {
    createManager = NewMockManager
}

var items = map[string]Account{
    "A": {1, "A", "USD", Cents(10000), time.Now()},
    "B": {2, "B", "USD", Cents(0), time.Now()},
    "C": {3,"C","EUR", Cents(1000), time.Now()},
}

type MockManager struct {
    Accounts map[string]Account
    Payments []Payment
}

func NewMockManager(_ string) (Manager, error) {
    var manager Manager = MockManager{items, make([]Payment, 0)}
    return manager, nil
}

func (m MockManager) Close() error { return nil }

func (m MockManager) GetAvailableAccounts() ([]Account, error) {
    accounts := make([]Account, 0)
    for _, acc := range m.Accounts {
        accounts = append(accounts, acc)
    }
    return accounts, nil
}

func (m MockManager) GetAccounts(identifiers []string) ([]Account, error) {
    filtered := make([]Account, 0)
    for _, id := range identifiers {
        filtered = append(filtered, m.Accounts[id])
    }
    return filtered, nil
}

func (m MockManager) Transfer(fromId, toId string, amount Cents) (*Payment, error) {
    first, ok := m.Accounts[fromId]
    if !ok {
        return nil, fmt.Errorf("fromId is missing")
    }

    second, ok := m.Accounts[toId]
    if !ok {
        return nil, fmt.Errorf("toId is missing")
    }

    if first.Amount < amount || first.Currency != second.Currency {
        return nil, fmt.Errorf("invalid configuration")
    }

    first.Amount -= amount
    second.Amount += amount
    payment := Payment{
        From:first.Identifier,
        To:second.Identifier,
        Time:time.Now().UTC(),
        Amount:amount,
        Currency:first.Currency}


    return &payment, nil
}

func (m MockManager) GetPayments(accountId string) ([]Payment, error) {
    payments := make([]Payment, 0)
    for _, p := range m.Payments {
        if p.From == accountId || p.To == accountId {
            payments = append(payments, p)
        }
    }
    return payments, nil
}