package main

import (
	"database/sql"
	"fmt"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
)

func dbConn() (db *sql.DB) {
	dbDriver := "mysql"
	dbUser := "root"
	dbPass := "password"
	dbName := "messenger"
	db, err := sql.Open(dbDriver, dbUser+":"+dbPass+"@/"+dbName)

	if err != nil {
		panic(err.Error())
	} else {
		fmt.Println("Successfully connected to mysql")
	}
	return db
}

// 0 means unauthorized else authorized

func getAuthID(w http.ResponseWriter, r *http.Request) int {
	sessionCookie, sessionErr := r.Cookie("session_id")
	if sessionErr != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return 0
	}
	sessionID := sessionCookie.Value
	db := dbConn()
	defer db.Close()
	qStr := fmt.Sprintf(`SELECT user_id FROM sessions WHERE session_id = "%s";`, sessionID)
	rows, qErr := db.Query(qStr)

	if qErr != nil {
		return 0
	}

	var userID int
	if rows.Next() {
		rows.Scan(&userID)
		return userID
	}
	return 0

}

func getUserID(username string) int {
	db := dbConn()
	defer db.Close()

	qStr := fmt.Sprintf(`SELECT id FROM users WHERE username = "%s";`, username)

	id, idErr := db.Query(qStr)
	if idErr != nil {
		return 0
	}
	var userID int
	if id.Next() {
		id.Scan(&userID)
		return userID
	}
	return 0

}

func getUserName(userID int) string {
	db := dbConn()
	defer db.Close()

	qStr := fmt.Sprintf(`SELECT username FROM users WHERE id = "%d";`, userID)

	id, idErr := db.Query(qStr)
	if idErr != nil {
		return ""
	}
	var username string
	if id.Next() {
		id.Scan(&username)
		return username
	}
	return ""
}
