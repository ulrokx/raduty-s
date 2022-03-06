# RA Duty Scheduler Server

This is the backend for the Resident Assistant scheduling tool.
It is written in Go 1.17. It uses GORM as an ORM and Gin as a
router/context manager.

## Setup Server

_this should ideally be set up with the
[frontend](https://github.com/ulrokx/raduty) here._

1. Download [Golang](https://go.dev/dl/), the default installer
   for whatever platform should work fine.
2. Download [PostgreSQL](https://www.postgresql.org/download/),
   this is required for the server to run.
3. Clone this repository with
   `git clone https://github.com/ulrokx/raduty-s.git`
4. Run `go mod download` in the main directory that has the go.mod file.
5. Run `go run ./main.go`

## To Do
- Algorithm in `/api/controllers/schedule.go` doesn't always fill the days up.
- Extract the main generation function out to test separately
- Write tests :-)
- Add camel case json tags to structs 
