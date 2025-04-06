package main

import (
	"encoding/gob"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"text/template"
	"time"

	"github.com/alexedwards/scs/mysqlstore"
	"github.com/alexedwards/scs/v2"
	"github.com/go-stripe/internal/driver"
	"github.com/go-stripe/internal/models"
)

const version = "v1"
const cssVersion = "v1"

var session scs.SessionManager

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
	DBModel       models.DBModel
	Session       *scs.SessionManager
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
	gob.Register(TransactionData{})
	var cfg config

	flag.IntVar(&cfg.port, "port", 4000, "define port")
	flag.StringVar(&cfg.api, "api", "http://localhost:4001", "define api endpoint")
	flag.StringVar(&cfg.env, "env", "development", "define env{development|production}")
	flag.StringVar(&cfg.db.dsn, "dsn", "root:secret@tcp(localhost:3306)/widgets?parseTime=true&tls=false", "define dsn")

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

	dbConnection, err := driver.OpenDB(app.config.db.dsn)

	app.DBModel = models.DBModel{
		DB: dbConnection,
	}

	if err != nil {
		app.errorLog.Fatal("Failed to connect to the database", err)
		return
	}

	defer dbConnection.Close()

	session = *scs.New()
	session.Lifetime = 24 * time.Hour
	session.Store = mysqlstore.New(dbConnection)

	app.Session = &session

	err = app.serve()

	if err != nil {
		app.errorLog.Panic("error starting server", err)
	}
}
