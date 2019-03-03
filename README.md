# Go Coding Challenge

A simple generic payment service implementation with a plain Go  and a couple of additional dependencies
to deal with the database. The project uses Docker Compose to manage containers and deploy API. 


## Dependencies

1. `go get github.com/lib/pq` - PostgreSQL driver written in Go. 
2. `go get github.com/jmoiron/sqlx` - A set of extensions for the standard `sql` package to make interactions 
with the database more convenient.


## Deployment

The `docker-compose.yml` file defines two containers:
1. `api`- a container with the REST API application
2. `db` - a container with the PostgreSQL database  

To deploy the app and make it ready to accept requests, it should be enough to install the Docker and run the
following commands:
```
$ docker-compose build
$ docker-compose up
``` 

The database contains two tables only, `account` and `payment`. The amount of money on the account is stored
as an integer number of "cents" to deal with possible rounding errors that can occur in case of floating-point numbers.

Note that having a database withing a container is probably not a strict requirement in production setting. The
database can be (and probably should be) deployed on a dedicated high-performance host. Also, we use a single
`docker-compose.yml` file while in general we should create separate configuration files for 
`test`, `stage`, `prod`, etc.

## Endpoints

As soon as the containers are up, we can start making HTTP requests. 
All examples use a handy [`httpie`](https://httpie.org) utility.

### `/status`

A testing endpoint to ping the API and check if the server is up.
```
$ http http://localhost:8080/status | jq .
{
  "success": true
}
```

### `/accounts`
Retrieves list of available accounts from the database.
```
$ http http://localhost:8080/accounts | jq .
{
  "accounts": [
    {
      "name": "first",
      "currency": "USD",
      "amount": "10.00"
    },
    {
      "name": "second",
      "currency": "USD",
      "amount": "0.00"
    },
    {
      "name": "third",
      "currency": "EUR",
      "amount": "0.10"
    }
  ]
}
``` 


### `/transfer`

Moves funds from one account to the another, but only if they have the same currency, and there is a sufficient
amount of funds.
```
http http://localhost:8080/transfer fromId=first toId=second amount=100 | jq .
{
  "payment": {
    "id": 0,
    "from": "first",
    "to": "second",
    "time_utc": "2019-03-03T08:30:53.039799678Z",
    "amount": 100,
    "currency": "USD"
  }
}
```

### `/payments`

Returns a list of transactions for the specific account.

```
$ http http://localhost:8080/payments accountId=first | jq .
{
  "account": "first",
  "payments": {
    "received": null,
    "sent": [
      {
        "account": "second",
        "amount": 1,
        "time": "2019-03-03T08:30:53.0398Z"
      }
    ]
  }
}
```

## Tests

The endpoint tests are stored in the file `api/src/server/server_test.go`. The tests use mockery to replace
PostgreSQL database queries with in-memory dataset. The tests don't provide the full coverage of the codebase 
but verify that endpoints work as expected with valid and invalid parameters.

To run tests, use the code: 
```
$ cd api 
$ go test -v ./src/server 
``` 