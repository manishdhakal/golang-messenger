package main

import (
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var jwtKey = []byte("manishjwt")

// UserCredentials for body of request
type UserCredentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// type Message struct {
// 	var Body string `json:"body"`
// }

type Conversation struct {
	RivalUsername string `json:"rivalUsername"`
}

func dbConn() (db *sql.DB) {
	dbDriver := "mysql"
	// dbUser := "root"
	// dbPass := "mpass074"
	// dbURL := "tcp(127.0.0.1:3306)/"
	// dbName := "messenger"
	db, err := sql.Open(dbDriver, "root:mpass074@tcp(127.0.0.1:3306)/messenger")

	if err != nil {
		panic(err.Error())
	} else {
		fmt.Println("Successfully connected to mysql")
	}
	return db
}

func main() {

	db := dbConn()
	db.Close()

	http.HandleFunc("/users", getUsers)

	http.HandleFunc("/logout", logoutHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/signup", createUser)

	http.HandleFunc("/create-conv", createConversation)
	fmt.Println("The server is running in port 8000......")
	log.Fatal(http.ListenAndServe("localhost:8000", nil))
}

// the function that returns all users
func getUsers(w http.ResponseWriter, r *http.Request) {

	db := dbConn()
	rows, err := db.Query("SELECT id,username FROM users")

	if err != nil {
		panic(err.Error())
	}
	fmt.Fprintf(w, "id\tusername")
	for rows.Next() {
		var id int
		var username string

		err = rows.Scan(&id, &username)

		fmt.Print(id, " ", username, "\n")
		fmt.Fprintf(w, fmt.Sprintf("\n%d\t", id))
		fmt.Fprintf(w, username)
	}
	defer db.Close()
}

// LoginHandler is used for authentication
func loginHandler(w http.ResponseWriter, r *http.Request) {

	var creds UserCredentials

	// Get the JSON body and decode into credentials
	jsonErr := json.NewDecoder(r.Body).Decode(&creds)
	if jsonErr != nil {
		// If the structure of the body is wrong, return an HTTP error
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	db := dbConn()
	defer db.Close()
	qStr := fmt.Sprintf(`SELECT EXISTS (
		SELECT * FROM users WHERE username = "%s" AND password_hash = SHA2("%s",256)
	  ) access`, creds.Username, creds.Password)

	rows, qErr := db.Query(qStr)

	if qErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Internal Server Error\n")
		return
	}

	var access int
	rows.Next()
	rows.Scan(&access)

	if access == 0 {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, "Invalid credentials\n")
		return
	}

	expTime := time.Now().AddDate(0, 2, 0)

	//create a hS 256 signer
	rawStr := fmt.Sprintf("%s%v%s", creds.Username, expTime, "manish_key")
	tkBuf := sha256.Sum256([]byte(rawStr))
	sessionID := fmt.Sprintf("%x", tkBuf)

	// save the token in the db
	expDay := fmt.Sprintf("%v-%d-%v", expTime.Year(), expTime.Month(), expTime.Day())

	sessQuery := fmt.Sprintf(`INSERT INTO sessions (session_id,user_id, expiry)
	VALUES ("%s", (SELECT id FROM users WHERE username="%s"), "%s");`, sessionID, creds.Username, expDay)

	_, sessErr := db.Query(sessQuery)

	if sessErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Internal Server Error\n")
		return
	}
	//set cookie in the browser
	http.SetCookie(w, &http.Cookie{
		Name:    "session_id",
		Value:   sessionID,
		Expires: expTime,
	})

}

// LogoutHandler to delete cookies and session
func logoutHandler(w http.ResponseWriter, r *http.Request) {

	// set cookie to be empty
	http.SetCookie(w, &http.Cookie{
		Name:  "session_id",
		Value: "",
	})

}

// to create conversation
func createConversation(w http.ResponseWriter, r *http.Request) {

	sessionCookie, sessionErr := r.Cookie("session_id")
	if sessionErr != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	sessionID := sessionCookie.Value

	var conv Conversation
	jsonErr := json.NewDecoder(r.Body).Decode(&conv)
	if jsonErr != nil {
		// If the structure of the body is wrong, return an HTTP error
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	db := dbConn()
	defer db.Close()

	if isAuthorized(sessionID) {
		var authUserID, rivalUserID int

		qStr := fmt.Sprintf(`SELECT user_id FROM sessions WHERE session_id = "%s"`, sessionID)
		id, idErr := db.Query(qStr)
		if idErr != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Print("Cannot retain your auth id\n")
			return
		}
		id.Next()
		id.Scan(&authUserID)

		qStr = fmt.Sprintf(`SELECT id FROM users WHERE username = "%s"`, conv.RivalUsername)
		rivalID, rivErr := db.Query(qStr)
		if rivErr != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Print("Cannot retain rival id\n")
			return
		}
		rivalID.Next()
		rivalID.Scan(&rivalUserID)

		qStr = fmt.Sprintf(`SELECT EXISTS (
			 	SELECT * FROM conversations WHERE (user1_id = %d AND user2_id = %d) OR (user1_id = %d AND user2_id = %d)
			) existence;`, authUserID, rivalUserID, rivalUserID, authUserID)
		existRow, existErr := db.Query(qStr)
		if existErr != nil {
			w.WriteHeader(409)
			fmt.Print("Error to check the validation of already existing conv\n")
			return
		}
		var exists int
		existRow.Next()
		existRow.Scan(&exists)

		if exists == 1 {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, "The conversation already exists\n")
			return
		}

		qStr = fmt.Sprintf(`INSERT INTO conversations (user1_id, user2_id)
				VALUES (%d, %d)`, authUserID, rivalUserID)
		_, InsErr := db.Query(qStr)
		if InsErr != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	} else {
		w.WriteHeader(http.StatusUnauthorized)
	}

}

// to create a user
func createUser(w http.ResponseWriter, r *http.Request) {
	var user UserCredentials
	// Get the JSON body and decode into credentials
	jsonErr := json.NewDecoder(r.Body).Decode(&user)
	if jsonErr != nil {
		// If the structure of the body is wrong, return an HTTP error
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	db := dbConn()
	defer db.Close()

	qStr := fmt.Sprintf(`SELECT EXISTS (
		SELECT * FROM users WHERE username = "%s"
	) exist;`, user.Username)
	exist, existErr := db.Query(qStr)
	if existErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Print(existErr.Error(), "\n")
		return
	}
	var exst int
	exist.Next()
	exist.Scan(&exst)

	if exst == 1 {
		w.WriteHeader(409)
		fmt.Print("The username already exists\n")
		return
	}

	qStr = fmt.Sprintf(`INSERT INTO users (username, password_hash)
			VALUES ("%s", SHA2("%s", 256))`, user.Username, user.Password)
	_, insErr := db.Query(qStr)

	if insErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Print(insErr.Error(), "\n")
		return
	}

	expTime := time.Now().AddDate(0, 2, 0)

	//create sha256 hex code
	rawStr := fmt.Sprintf("%s%v%s", user.Username, expTime, "manish_key")
	tkBuf := sha256.Sum256([]byte(rawStr))
	sessionID := fmt.Sprintf("%x", tkBuf)

	// save the token in the db
	expDay := fmt.Sprintf("%v-%d-%v", expTime.Year(), expTime.Month(), expTime.Day())

	sessQuery := fmt.Sprintf(`INSERT INTO sessions (session_id,user_id, expiry)
	VALUES ("%s", (SELECT id FROM users WHERE username="%s"), "%s");`, sessionID, user.Username, expDay)

	_, sessErr := db.Query(sessQuery)

	if sessErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Internal Server Error\n")
		return
	}
	//set cookie in the browser
	http.SetCookie(w, &http.Cookie{
		Name:    "session_id",
		Value:   sessionID,
		Expires: expTime,
	})

}

func isAuthorized(sessionID string) bool {
	db := dbConn()
	defer db.Close()
	qStr := fmt.Sprintf(`SELECT EXISTS (
		SELECT * FROM sessions WHERE session_id = "%s"
	) auth;`, sessionID)
	rows, qErr := db.Query(qStr)

	if qErr != nil {
		return false
	}

	var access int
	rows.Next()
	rows.Scan(&access)

	if access == 0 {
		return false
	}

	return true

}
