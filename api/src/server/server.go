package server

import (
    "encoding/json"
    "fmt"
    "io"
    "log"
    "net/http"
    "strconv"
)

type Response map[string]interface{}

type Config struct {
    Host string
    Port int
    DatabaseConn string
}

func (c Config) Addr() string { return fmt.Sprintf("%s:%d", c.Host, c.Port) }
func (c Config) URL()  string { return fmt.Sprintf("http://%s", c.Addr()) }

type BillingAPI struct {
    Config
    *http.Server
}

func NewBillingAPI(conf Config) *BillingAPI {
    mux := http.NewServeMux()
    api := BillingAPI{conf, &http.Server{Addr:conf.Addr(), Handler:mux}}
    mux.Handle("/", http.HandlerFunc(notFound))
    mux.Handle("/status", http.HandlerFunc(api.status))
    mux.Handle("/accounts", http.HandlerFunc(api.accounts))
    mux.Handle("/transfer", http.HandlerFunc(api.transfer))
    return &api
}

func (api *BillingAPI) status(w http.ResponseWriter, req *http.Request) {
    log.Printf("connected: %s", req.RemoteAddr)
    resp := NewJSONResponse(w)
    resp.SendSuccess(Response{"success": true})
}

func (api *BillingAPI) accounts(w http.ResponseWriter, req *http.Request) {
    resp := NewJSONResponse(w)

    m, err := NewManager(api.DatabaseConn)
    if err != nil {
        log.Println(err)
        resp.SendServerError("internal error")
        return
    } else {
        defer closeWithLog(m)
    }

    accounts, err := m.GetAvailableAccounts()
    if err != nil {
        log.Println(err)
        resp.SendServerError("internal error")
        return
    }

    type info struct {
        Name string     `json:"name"`
        Currency string `json:"currency"`
    }
    result := make([]info, 0)
    for _, acc := range accounts {
        result = append(result, info{acc.Identifier, acc.Currency})
    }
    resp.SendSuccess(Response{"accounts": result})
}

// transfer endpoint moves specified amount of funds from one account to another.
//
// The endpoint expects the following parameters:
//     * fromId: an account where to take the money
//     * toId: an account where to send the money
//     * amount: an amount of money (as an integer number of cents) to transfer
//
// Example of possible request's body:
//
//     {"fromId": "account_1", "toId": "account_2", "amount": 1000}
//
// Note that only transfer between accounts with the same currency is supported.
// In case if any of accounts doesn't exist, if there is no enough funds, or
// the currency of accounts is different, the error is returned.
func (api *BillingAPI) transfer(w http.ResponseWriter, req *http.Request) {
    resp := NewJSONResponse(w)

    m, err := NewManager(api.DatabaseConn)
    if err != nil {
        log.Println(err)
        resp.SendServerError("internal error")
        return
    } else {
        defer closeWithLog(m)
    }

    data := make(map[string]string)
    err = json.NewDecoder(req.Body).Decode(&data)
    if err != nil {
        resp.SendRequestError("invalid request body")
        return
    }

    err = CheckParameters(data, "from", "to", "amount")
    if err != nil {
        resp.SendRequestError(fmt.Sprintf("invalid request: %s", err))
        return
    }

    amount, err := strconv.Atoi(data["amount"])
    if err != nil {
        resp.SendRequestError("invalid amount value")
        return
    }

    fromId, toId := data["from"], data["to"]
    payment, err := m.Transfer(fromId, toId, Cents(amount))
    if err != nil {
        // do not disclose internal error to the client
        if err, ok := err.(managerError); ok {
            if err.internal {
                log.Printf("error: %s", err.message)
                resp.SendServerError("internal error")
            } else {
                resp.SendRequestError(err.Error())
            }
        } else {
            resp.SendError(err, http.StatusBadRequest)
        }
        return
    }

    resp.SendSuccess(Response{"payment": payment})
}

func notFound(w http.ResponseWriter, req *http.Request) {
    NewJSONResponse(w).SendRequestError("not found")
}

func closeWithLog(c io.Closer) {
    err := c.Close()
    if err != nil {
        log.Printf("warning: error on closing object %v: %s", c, err)
    }
}