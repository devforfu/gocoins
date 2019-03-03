package server

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "sync"
    "testing"
    "time"
)

func init() {
    createManager = NewMockManager
}

func TestAccounts(t *testing.T) {
    makeRequest(t, func(client TestClient) {
        result := client.JSONRequest("GET", "accounts", nil)
        if value, ok := result["accounts"]; !ok {
            t.Errorf("no 'accounts' key found: %#v", result)
        } else {
            responseItems := value.([]interface{})
            if len(responseItems) != len(items) {
                t.Errorf("invalid number of results")
            }
        }
    })
}

func TestTransfer_ValidParameters(t *testing.T) {
    makeRequest(t, func(client TestClient) {
        result := client.JSONRequest("GET", "transfer", map[string]string{
            "fromId": "A",
            "toId": "B",
            "amount": "1000",
        })
        if value, ok := result["payment"]; !ok {
            log.Println(result)
            t.Errorf("no 'payment' key found: %#v", result)
        } else {
            responseItems := value.(map[string]interface{})
            for _, key := range []string{"from", "to", "time_utc", "amount", "currency"} {
                _, ok = responseItems[key]
                if !ok {
                    t.Errorf("key is missing: %s", key)
                }
            }
            if value, ok := responseItems["amount"].(float64); !ok {
                t.Errorf("invalid format: %s", responseItems["amount"])
            } else {
                if value != 1000 {
                    t.Errorf("invalid amount: %f != 1000", value)
                }
            }
        }
    })
}

func TestTransfer_InvalidParameters(t *testing.T) {
   makeRequest(t, func(client TestClient) {
       var testCases = []map[string]string {
           {"fromId": "X", "toId": "Z", "amount": "1000"},
           {"fromId": "A", "toId": "B", "amount": "0"},
           {"fromId": "A", "toId": "B", "amount": "-1"},
           {"fromId": "A", "toId": "B", "amount": "999999"},
           {"fromId": "A", "toId": "C", "amount": "1000"},
       }
       for _, params := range testCases {
           response := client.JSONRequest("GET", "transfer", params)
           _, ok := response["error"]
           if !ok {
               t.Errorf("error was expected for params: %v", params)
           }
       }
   })
}

func TestPayments_ValidParameters(t *testing.T) {
   makeRequest(t, func(client TestClient) {
       var testCases = []struct{
           accountId string
           nFrom, nTo int
       }{
           {"A", 1, 1},
           {"B", 1, 1},
           {"C", 0, 0},
       }
       for _, test := range testCases {
           params := map[string]string{"accountId": test.accountId}
           response := client.JSONRequest("GET", "payments", params)

           accountId, ok := response["account"]
           if !ok { t.Errorf("missing key 'account'") }

           payments, ok := response["payments"]
           if !ok { t.Errorf("missing key 'payments'") }

           if accountId.(string) != test.accountId { t.Errorf("invalid accountId") }

           paymentsDict := payments.(map[string]interface{})

           _, ok = paymentsDict["sent"]
           if !ok { t.Error("key 'sent' is not found") }

           _, ok = paymentsDict["received"]
           if !ok { t.Error("key 'received' is not found") }

           if (test.nFrom == 0 && paymentsDict["sent"] != nil) ||
              (test.nFrom > 1 && len(paymentsDict["sent"].([]interface{})) != test.nFrom) ||
              (test.nTo == 0 && paymentsDict["received"] != nil) ||
              (test.nTo > 1 && len(paymentsDict["received"].([]interface{})) != test.nTo) {
               t.Error("expected and received results don't match")
           }
       }
   })
}

func TestPayments_InvalidParameters(t *testing.T) {
   makeRequest(t, func(client TestClient) {
       params := map[string]string{"accountId": "Unknown"}
       response := client.JSONRequest("GET", "payments", params)
       if _, ok := response["error"]; !ok {
           t.Errorf("error was expected: %#v", response)
       }
   })
}


// -----------
// Test client
// -----------


func makeRequest(t *testing.T, testCase func(client TestClient)) {
    api := NewBillingAPI(Config{"",8080,""})
    group := sync.WaitGroup{}
    group.Add(1)

    go func() {
        if err := api.ListenAndServe(); err != http.ErrServerClosed {
            t.Errorf("server error: %s", err)
        }
        group.Done()
    }()

    testCase(TestClient{"http://localhost:8080", t})
    _ = api.Shutdown(context.TODO())
    group.Wait()
}

type TestClient struct {
    Schema string
    Test *testing.T
}

func (c *TestClient) JSONRequest(method, endpoint string, query map[string]string) Response {
    url := c.URL(endpoint)
    encoded, _ := json.Marshal(query)
    req, err := http.NewRequest(method, url, bytes.NewBuffer(encoded))
    if err != nil { c.Test.Error(err) }

    req.Header.Set("Content-Type", "application/json")
    client := http.Client{}
    resp, err := client.Do(req)
    if err != nil { c.Test.Error(err) }
    defer resp.Body.Close()

    var result Response
    err = json.NewDecoder(resp.Body).Decode(&result)
    if err != nil { c.Test.Error(err) }
    return result
}

func (c *TestClient) URL(endpoint string) string {
    return fmt.Sprintf("%s/%s", c.Schema, endpoint)
}


// ---------------------
// Manager class mockery
// ---------------------


var items = map[string]Account{
    "A": {1, "A", "USD", Cents(10000), time.Now()},
    "B": {2, "B", "USD", Cents(1000), time.Now()},
    "C": {3, "C", "EUR", Cents(5000), time.Now()}}

var payments = []Payment{
    {1, "A", "B", time.Now().Add(-1*time.Hour), 1000, "USD"},
    {2, "B", "A", time.Now().Add(-2*time.Hour), 1000, "USD"}}


// A MockManager type replaces real database management with mock implementation.
// The MockManager uses two in-memory arrays, Accounts and Payments, with the predefined data.
// It doesn't store the performed changes and expected to be stateless.
type MockManager struct {
    Accounts map[string]Account
    Payments []Payment
}

func NewMockManager(_ string) (Manager, error) {
    var manager Manager = MockManager{items, payments}
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
    if _, ok := m.Accounts[accountId]; !ok {
        return nil, fmt.Errorf("account is not found")
    }
    for _, p := range m.Payments {
        if p.From == accountId || p.To == accountId {
            payments = append(payments, p)
        }
    }
    return payments, nil
}
