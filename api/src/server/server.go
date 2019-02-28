package server

import (
    "encoding/json"
    "fmt"
    "log"
    "net/http"
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
    return &api
}

func (api *BillingAPI) status(w http.ResponseWriter, req *http.Request) {
    log.Printf("connected: %s", req.RemoteAddr)
    resp := NewJSONResponse(w)
    resp.SendSuccess(Response{"success": true})
}

func (api *BillingAPI) accounts(w http.ResponseWriter, req *http.Request) {
    resp := NewJSONResponse(w)

    conn, err := connect(api.DatabaseConn)
    if err != nil {
        log.Print(err)
        resp.SendServerError("internal error")
        return
    }

    rows, err := conn.Query("SELECT * FROM account")
    if err != nil {
        log.Print(err)
        resp.SendServerError("internal error")
        return
    }

    accounts := make([]Account, 0)
    for rows.Next() {
        var account Account
        err := rows.Scan(&account.ID, &account.Identifier, &account.Currency, &account.Created)
        if err != nil {
            log.Printf("fetch error: %s", err)
            continue
        }
        accounts = append(accounts, account)
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

    err := CheckParameters()
}

func notFound(w http.ResponseWriter, req *http.Request) {
    NewJSONResponse(w).SendRequestError("not found")
}
