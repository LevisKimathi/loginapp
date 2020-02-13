package main

import (
	"context"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/sessions"
	_ "github.com/gorilla/sessions"
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

var (
	key   = []byte("leviskimathi")
	store = sessions.NewCookieStore(key)
)

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
	s.mux.HandleFunc("/dashboard/logout", s.logout)

	return s
}

func (s *server) index(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "cookie-name")

	tmpl := template.Must(template.ParseFiles("index.html"))
	// Check if user is authenticated
	if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
		if r.Method == http.MethodPost {
			db := dbConn()
			selDB, err := db.Query("SELECT * FROM users WHERE username = ? AND password = ? ", r.FormValue("username"), r.FormValue("password"))
			if err != nil {
				panic(err.Error())
			}

			for selDB.Next() {
				var id int
				var username, email, phone, password string
				err = selDB.Scan(&id, &username, &email, &phone, &password)
				// Set user as authenticated
				session.Values["authenticated"] = true
				session.Values["username"] = username
				_ = session.Save(r, w)
				http.Redirect(w, r, "/dashboard", 301)
				if err != nil {
					panic(err.Error())
				}
			}

			log.Println("SELECTED")

			defer db.Close()

		}

		_ = tmpl.Execute(w, nil)
	} else {
		http.Redirect(w, r, "/dashboard", 301)
	}
}

func (s *server) register(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "cookie-name")

	tmpl := template.Must(template.ParseFiles("register.html"))

	// Check if user is authenticated
	if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
		if r.Method == http.MethodPost {

			db := dbConn()
			insForm, err := db.Prepare("INSERT INTO users(username, email, phone,  password) VALUES(?,?,?,?)")
			if err != nil {
				panic(err.Error())
			}

			username := r.FormValue("username")
			email := r.FormValue("email")
			phone := r.FormValue("phone")
			password := r.FormValue("password")

			_, _ = insForm.Exec(username, email, phone, password)

			log.Println("INSERTED")

			defer db.Close()

			http.Redirect(w, r, "/", 301)

		}
		_ = tmpl.Execute(w, nil)
	} else {
		http.Redirect(w, r, "/dashboard", 301)
	}
}

func (s *server) reset(w http.ResponseWriter, r *http.Request) {
	//Get session
	session, _ := store.Get(r, "cookie-name")

	tmpl := template.Must(template.ParseFiles("reset.html"))

	// Check if user is authenticated
	if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
		if r.Method == http.MethodPost {

		}
		_ = tmpl.Execute(w, nil)
	} else {
		http.Redirect(w, r, "/dashboard", 301)
	}
}

func (s *server) dashboard(w http.ResponseWriter, r *http.Request) {
	//Get session
	session, _ := store.Get(r, "cookie-name")

	tmpl := template.Must(template.ParseFiles("dashboard.html"))

	// Check if user is authenticated
	if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
		http.Redirect(w, r, "/", 301)
	} else {
		type required struct {
			Success  bool
			Username interface{}
		}

		_ = tmpl.Execute(w, required{true, session.Values["username"]})
	}
}

func (s *server) logout(w http.ResponseWriter, r *http.Request) {
	//Get session
	session, _ := store.Get(r, "cookie-name")

	// Revoke users authentication
	session.Values["authenticated"] = false
	_ = session.Save(r, w)

	http.Redirect(w, r, "/", 301)
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
