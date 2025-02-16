package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"text/template"
	"time"
)

const version = "v1"
const cssVersion = "v1"

type config struct {
	port int
	api  string
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
	config        config
	infoLog       log.Logger
	errorLog      log.Logger
	templateCache map[string]*template.Template
	version       string
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

	app.infoLog.Printf("application running on port %v in %v env\n mode", app.config.port, app.config.env)

	return server.ListenAndServe()
}

func main() {
	var cfg config

	flag.IntVar(&cfg.port, "port", 4000, "define port")
	flag.StringVar(&cfg.api, "api", "http://localhost:4001", "define api endpoint")
	flag.StringVar(&cfg.env, "env", "development", "define env{development|production}")

	flag.Parse()

	cfg.stripe.key = "pk_test_51QqX3TLNGyaF79S2XZ6vSEspSBCJDZ3A5NLkjTAQdMgePTXe7JEcyLGkinfbXDr1RvWGrOeRza7nYrymnm4zrDwA004QGGj7sT"
	cfg.stripe.secret = "sk_test_51QqX3TLNGyaF79S20fYOTYJRUcghafnpPpiz7zfmyoYafFuh9lf1ufIBRBLxda0DyYaTdALCeY5ESmDl59pg8zkH00ReJa86RT"

	app := &application{
		config:        cfg,
		infoLog:       *log.New(os.Stdout, "INFO:\t", log.Ldate|log.Ltime),
		errorLog:      *log.New(os.Stdout, "ERROR:\t", log.Ldate|log.Ltime|log.Lshortfile),
		templateCache: make(map[string]*template.Template),
		version:       version,
	}

	err := app.serve()

	if err != nil {
		app.errorLog.Panic("error starting server", err)
	}
}
