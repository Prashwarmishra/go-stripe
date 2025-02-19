package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-stripe/internal/driver"
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
	flag.StringVar(&cfg.db.dsn, "dsn", "prashwarm:secret@tcp(localhost:3306)/widgets?parseTime=true&tls=false", "define dsn")

	flag.Parse()

	// cfg.stripe.key = os.Getenv("STRIPE_KEY")
	// cfg.stripe.secret = os.Getenv("STRIPE_SECRET")

	cfg.stripe.key = "pk_test_51QqX3TLNGyaF79S2XZ6vSEspSBCJDZ3A5NLkjTAQdMgePTXe7JEcyLGkinfbXDr1RvWGrOeRza7nYrymnm4zrDwA004QGGj7sT"
	cfg.stripe.secret = "sk_test_51QqX3TLNGyaF79S20fYOTYJRUcghafnpPpiz7zfmyoYafFuh9lf1ufIBRBLxda0DyYaTdALCeY5ESmDl59pg8zkH00ReJa86RT"

	app := &application{
		config:   cfg,
		infoLog:  log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime),
		errorLog: log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime|log.Lshortfile),
		version:  version,
	}

	dbConnection, err := driver.OpenDB(app.config.db.dsn)

	if err != nil {
		app.errorLog.Fatal("Failed to connect to the backend database", err)
		return
	}

	defer dbConnection.Close()

	err = app.serve()

	if err != nil {
		app.errorLog.Println("error in starting backend server", err)
	}
}
