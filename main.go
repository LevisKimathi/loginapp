package main

import (
	"context"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func dbConn() (db *sql.DB) {
	dbDriver := "mysql"
	dbUser := "root"
	dbPass := ""
	dbName := "goproject"
	db, err := sql.Open(dbDriver, dbUser+":"+dbPass+"@/"+dbName)
	if err != nil {
		panic(err.Error())
	}
	return db
}

type server struct {
	logger *log.Logger
	mux    *http.ServeMux
}

func newServer(options ...func(*server)) *server {
	s := &server{mux: http.NewServeMux()}

	for _, f := range options {
		f(s)
	}

	if s.logger == nil {
		s.logger = log.New(os.Stdout, "", 0)
	}

	s.mux.HandleFunc("/", s.index)
	s.mux.HandleFunc("/register/", s.register)
	s.mux.HandleFunc("/reset/", s.reset)
	s.mux.HandleFunc("/dashboard/", s.dashboard)

	return s
}

func (s *server) index(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("index.html"))
	if r.Method == http.MethodPost {

		db := dbConn()
		selDB, err := db.Query("SELECT * FROM users ORDER BY id DESC")
		if err != nil {
			panic(err.Error())
		}
		for selDB.Next() {
			var id int
			var username, email, phone, password string
			err = selDB.Scan(&id, &username, &email,  &phone, &password)
			http.Redirect(w, r, "/dashboard", 301)
			if err != nil {
				panic(err.Error())
			}
		}

		log.Println("SELECTED")

		defer db.Close()

	}
	tmpl.Execute(w, nil)
}

func (s *server) register(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("register.html"))
	if r.Method == http.MethodPost {

		db := dbConn()
		insForm, err := db.Prepare("INSERT INTO users(username, email, phone,  password) VALUES(?,?,?,?)")
		if err != nil {
			panic(err.Error())
		}

		username :=  r.FormValue("username")
		email :=  r.FormValue("email")
		phone :=  r.FormValue("phone")
		password :=  r.FormValue("password")

		insForm.Exec(username, email, phone, password)

		log.Println("INSERTED")

		defer db.Close()

		http.Redirect(w, r, "/", 301)

	}
	tmpl.Execute(w, nil)
}

func (s *server) reset(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("reset.html"))
	if r.Method == http.MethodPost {

	}
	tmpl.Execute(w, nil)
}

func (s *server) dashboard(w http.ResponseWriter, r *http.Request,) {
	tmpl := template.Must(template.ParseFiles("dashboard.html"))
	tmpl.Execute(w, struct{ Success bool }{true})
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Server", "Go server")

	s.mux.ServeHTTP(w, r)
}

func main() {
	hs, logger := setup()

	go func() {
		logger.Printf("Listening on http://127.0.0.1%s\n", hs.Addr)

		if err := hs.ListenAndServe(); err != http.ErrServerClosed {
			logger.Fatal(err)
		}
	}()

	graceful(hs, logger, 5*time.Second)
}

func setup() (*http.Server, *log.Logger) {
	addr := ":" + os.Getenv("PORT")
	if addr == ":" {
		addr = ":8000"
	}

	logger := log.New(os.Stdout, "", 0)

	s := newServer(func(s *server) {
		s.logger = logger
	})

	hs := &http.Server{Addr: addr, Handler: s}

	return hs, logger
}

func graceful(hs *http.Server, logger *log.Logger, timeout time.Duration) {
	stop := make(chan os.Signal, 1)

	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	logger.Printf("\nShutdown with timeout: %s\n", timeout)

	if err := hs.Shutdown(ctx); err != nil {
		logger.Printf("Error: %v\n", err)
	} else {
		logger.Println("Server stopped")
	}
}

