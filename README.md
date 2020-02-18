# loginapp
Simple go lang project that demonstrates how to build a login, registration and reset

Implements
1. Graceful shutdown
2. Templates
3. Encryption
4. Sessions
5. Server
6. Gvt for vendoring
7. Heroku for hosting
# Setup Local

For set up on your machine .

    Clone the repo git clone https://github.com/LevisKimathi/loginapp.git.
    Run go mod init to check if go modules is already initialized.
    Touch main.go file and paste the following configurations.

## main.go
```go
func dbConn() (db *sql.DB) {
	dbDriver := "mysql"
	dbUser := "username" # <-provide your own-->
	dbPass := "db_password" # <-provide your own-->
	dbName := "db_name" # <-provide your own-->
	//Create a new mysql db connection
	db, err := sql.Open(dbDriver, dbUser+":"+dbPass+"@/"+dbName)
	if err != nil {
		panic(err.Error())
	}
	return db
}
```
## Database 
    Open the .sql file in the database directory.
    Open your localhost and import the sql file into your database
## Run
     Go run 'go run main.go' on your terminal in the project DIR
     Open your browser and move to the specified route http://127.0.0.1:8000
# Todo

 - [ ] Password Reset with email
 - [ ] Dashboard
 - [ ] Testing
# Technologies Used

Here's a list of technologies used in this project

  * [Golang version go1.13](https://golang.org/doc/go1.13)  
  * [go-sql-driver/mysql](https://github.com/go-sql-driver/mysql)  
  * [gorilla/sessions](https://github.com/gorilla/sessions)  
  * [crypto/bcrypt](https://golang.org/x/crypto/bcrypt)  