package main

import (
	"context"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/sessions"
	"golang.org/x/crypto/bcrypt"
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
	//Create a new mysql db connection
	db, err := sql.Open(dbDriver, dbUser+":"+dbPass+"@/"+dbName)
	if err != nil {
		panic(err.Error())
	}
	return db
}

var (
	// Encrypt cookies using a key
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

	//Handle web routes
	s.mux.HandleFunc("/", s.index)
	s.mux.HandleFunc("/register", s.register)
	s.mux.HandleFunc("/reset", s.reset)
	s.mux.HandleFunc("/dashboard", s.dashboard)
	s.mux.HandleFunc("/dashboard/logout", s.logout)

	//Handle static server files such as images,css,js e.t.c
	s.mux.Handle("/static/img/",
		http.StripPrefix("/static/img/", http.FileServer(http.Dir("./static/img/"))))
	return s
}

func (s *server) index(w http.ResponseWriter, r *http.Request) {
	// Get session
	session, _ := store.Get(r, "cookie-name")

	// Create template from html file
	tmpl := template.Must(template.ParseFiles("templates/index.tmpl.html", "templates/header.tmpl.html", "templates/footer.tmpl.html"))
	// Check if user is authenticated
	if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
		if r.Method == http.MethodPost {
			password := r.FormValue("password")

			db := dbConn()
			selDB, err := db.Query("SELECT * FROM users WHERE username = ?", r.FormValue("username"))
			if err != nil {
				panic(err.Error())
			}

			for selDB.Next() {
				var id int
				var username, email, phone, hash string
				err = selDB.Scan(&id, &username, &email, &phone, &hash)

				if err != nil {
					panic(err.Error())
				}

				// Match hash in db with current hash
				pwdMatch := comparePasswords(hash, []byte(password))

				if pwdMatch {
					// Set user as authenticated
					session.Values["authenticated"] = true
					session.Values["username"] = username
					_ = session.Save(r, w)
					http.Redirect(w, r, "/dashboard", 301)

				}

			}

			log.Println("SELECTED")

			defer db.Close()

		}

		_ = tmpl.Execute(w, nil)
	} else {
		// Redirect to dashboard
		http.Redirect(w, r, "/dashboard", 301)
	}
}

func (s *server) register(w http.ResponseWriter, r *http.Request) {
	// Get session
	session, _ := store.Get(r, "cookie-name")

	// Create template from html file
	tmpl := template.Must(template.ParseFiles("templates/register.tmpl.html", "templates/header.tmpl.html", "templates/footer.tmpl.html"))

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
			// Hash the password
			hash := hashAndSalt([]byte(password))
			// Insert data into table
			_, _ = insForm.Exec(username, email, phone, hash)

			log.Println("INSERTED")

			defer db.Close()

			http.Redirect(w, r, "/", 301)

		}
		_ = tmpl.Execute(w, nil)
	} else {
		// Redirect to dashboard
		http.Redirect(w, r, "/dashboard", 301)
	}
}

func (s *server) reset(w http.ResponseWriter, r *http.Request) {
	// Get session
	session, _ := store.Get(r, "cookie-name")

	// Create template from html file
	tmpl := template.Must(template.ParseFiles("templates/reset.tmpl.html", "templates/header.tmpl.html", "templates/footer.tmpl.html"))

	// Check if user is authenticated
	if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
		if r.Method == http.MethodPost {

		}
		_ = tmpl.Execute(w, nil)
	} else {
		// Redirect to dashboard
		http.Redirect(w, r, "/dashboard", 301)
	}
}

func (s *server) dashboard(w http.ResponseWriter, r *http.Request) {
	// Get session
	session, _ := store.Get(r, "cookie-name")

	// Create template from html file
	tmpl := template.Must(template.ParseFiles("templates/dashboard.tmpl.html", "templates/header.tmpl.html", "templates/footer.tmpl.html"))

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
	// Get session
	session, _ := store.Get(r, "cookie-name")

	// Delete all sessions
	session.Options.MaxAge = -1
	// Revoke users authentication
	session.Values["authenticated"] = false
	_ = session.Save(r, w)

	// Redirect to default endpoint
	http.Redirect(w, r, "/", 301)
}

func hashAndSalt(pwd []byte) string {

	// Use GenerateFromPassword to hash & salt pwd.
	// MinCost is just an integer constant provided by the bcrypt package along with DefaultCost & MaxCost.
	// The cost can be any value you want provided it isn't lower than the MinCost (4)
	hash, err := bcrypt.GenerateFromPassword(pwd, bcrypt.MinCost)
	if err != nil {
		log.Println(err)
	}
	// GenerateFromPassword returns a byte slice so we need to convert the bytes to a string and return it
	return string(hash)
}

func comparePasswords(hashedPwd string, plainPwd []byte) bool {
	// Since we'll be getting the hashed password from the DB it will be a string so we'll need to convert it to a byte slice
	byteHash := []byte(hashedPwd)
	err := bcrypt.CompareHashAndPassword(byteHash, plainPwd)
	if err != nil {
		log.Println(err)
		return false
	}

	return true
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
