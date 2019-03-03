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
    mux.Handle("/payments", http.HandlerFunc(api.payments))
    return &api
}

var createManager = NewBillingManager

func (api *BillingAPI) status(w http.ResponseWriter, req *http.Request) {
    log.Printf("connected: %s", req.RemoteAddr)
    resp := NewJSONResponse(w)
    resp.SendSuccess(Response{"success": true})
}

func (api *BillingAPI) accounts(w http.ResponseWriter, req *http.Request) {
    resp := NewJSONResponse(w)

    m, err := createManager(api.DatabaseConn)
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
        Amount string   `json:"amount"`
    }
    result := make([]info, 0)
    for _, acc := range accounts {
        amount := fmt.Sprintf("%.2f", acc.Amount.AsFloat())
        result = append(result, info{acc.Identifier, acc.Currency, amount})
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

    m, err := createManager(api.DatabaseConn)
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

    err = CheckParameters(data, "fromId", "toId", "amount")
    if err != nil {
        resp.SendRequestError(fmt.Sprintf("invalid request: %s", err))
        return
    }

    amount, err := strconv.Atoi(data["amount"])
    if err != nil || amount <= 0 {
        resp.SendRequestError("invalid amount value")
        return
    }

    fromId, toId := data["fromId"], data["toId"]
    payment, err := m.Transfer(fromId, toId, Cents(amount))
    if err != nil {
        writeManagerError(err, &resp);
        return
    }

    resp.SendSuccess(Response{"payment": payment})
}

// payments endpoint reports transactions performed with a specific account.
//
// The endpoint expects the following parameters:
//     * accountId: an account which transactions to report.
//
// The error is returned in case if the account doesn't exist.
func (api *BillingAPI) payments(w http.ResponseWriter, req *http.Request) {
    resp := NewJSONResponse(w)

    m, err := createManager(api.DatabaseConn)
    if err != nil {
        log.Println(err)
        resp.SendServerError("internal error")
        return
    } else {
        defer closeWithLog(m)
    }

    data, err := decodeBody(req)
    if err != nil {
        resp.SendRequestError(fmt.Sprintf("invalid request: %s", err))
        return
    }

    accountId, ok := data["accountId"]
    if !ok {
        resp.SendRequestError("invalid request: accountId is missing")
        return
    }

    payments, err := m.GetPayments(accountId)
    if err != nil {
        writeManagerError(err, &resp)
        return
    }

    var send, recv []map[string]interface{}

    for _, p := range payments {
        item := make(map[string]interface{}, 0)
        item["amount"] = p.Amount.AsFloat()
        item["time"] = p.Time
        if p.From == accountId {
            item["account"] = p.To
            send = append(send, item)
        } else {
            item["account"] = p.From
            recv = append(recv, item)
        }
    }

    transactions := map[string]interface{}{"sent": send, "received": recv}
    resp.SendSuccess(Response{"account": accountId, "payments": transactions})
}

func notFound(w http.ResponseWriter, req *http.Request) {
    NewJSONResponse(w).SendRequestError("not found")
}

func decodeBody(req *http.Request) (map[string]string, error) {
    data := make(map[string]string)
    err := json.NewDecoder(req.Body).Decode(&data)
    if err != nil { return nil, err }
    _ = req.Body.Close()
    return data, nil
}

func closeWithLog(c io.Closer) {
    err := c.Close()
    if err != nil {
        log.Printf("warning: error on closing object %v: %s", c, err)
    }
}

// writeManagerError sends error response to the client. In case if the error
// comes from the invalid input, it is reported to the client. Otherwise, only
// a generic message about internal error is sent.
func writeManagerError(err error, resp *Responder) {
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
}