package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

const version = "v1"

type config struct {
	port int
	env  string
	db   struct {
		dsn string
	}
	stripe struct {
		secret string
		key    string
	}
}

type application struct {
	config   config
	infoLog  *log.Logger
	errorLog *log.Logger
	version  string
}

func (app *application) serve() error {
	server := http.Server{
		Addr:              fmt.Sprintf(":%d", app.config.port),
		Handler:           app.routes(),
		IdleTimeout:       30 * time.Second,
		ReadTimeout:       10 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      5 * time.Second,
	}

	app.infoLog.Printf("backend server running on port %v in %v env", app.config.port, app.config.env)
	return server.ListenAndServe()
}

func main() {
	var cfg config

	flag.IntVar(&cfg.port, "port", 4001, "define port to run backend server")
	flag.StringVar(&cfg.env, "env", "development", "define environment to run server {production|staging|maintainance}")

	flag.Parse()

	cfg.stripe.key = os.Getenv("STRIPE_KEY")
	cfg.stripe.secret = os.Getenv("STRIPE_SECRET")

	app := &application{
		config:   cfg,
		infoLog:  log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime),
		errorLog: log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime|log.Lshortfile),
		version:  version,
	}

	err := app.serve()

	if err != nil {
		app.errorLog.Println("error in starting backend server", err)
	}
}
