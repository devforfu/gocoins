package server

import (
    "encoding/json"
    "fmt"
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

    cur, err := NewCursor(api.DatabaseConn)
    if err != nil {
        log.Print(err)
        resp.SendServerError("internal error")
        return
    }
    defer cur.Close()

    accounts, err := cur.GetAvailableAccounts()
    if err != nil {
        log.Print(err)
        resp.SendServerError("internal error")
        return
    }

    resp.SendSuccess(Response{"accounts": accounts})
}

func (api *BillingAPI) transfer(w http.ResponseWriter, req *http.Request) {
    resp := NewJSONResponse(w)

    data := make(map[string]string)
    err := json.NewDecoder(req.Body).Decode(&data)
    if err != nil {
        resp.SendRequestError("invalid request body")
        return
    }

    err = CheckParameters(data, "from", "to", "amount")
    if err != nil {
        resp.SendRequestError(fmt.Sprintf("invalid request: %s", err))
        return
    }

    cur, err := NewCursor(api.DatabaseConn)
    if err != nil {
        log.Print(err)
        resp.SendServerError("internal error")
        return
    }
    defer cur.Close()

    fromId, toId := data["from"], data["to"]
    amount, err := strconv.Atoi(data["amount"])
    if err != nil {
        log.Print(err)
        resp.SendRequestError(fmt.Sprintf("invalid transfer amount: %s", data["amount"]))
        return
    }

    currency, err := cur.Transfer(fromId, toId, cents(amount))
    if err != nil {
        log.Print(err)
        resp.SendRequestError(err.Error())
        return
    }

    resp.SendSuccess(Response{"from": fromId, "to": toId, "currency": currency, "amount": cents(amount).String()})
}

func notFound(w http.ResponseWriter, req *http.Request) {
    NewJSONResponse(w).SendRequestError("not found")
}
