package main

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

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

// 0 means unauthorized else authorized
func getAuthID(sessionID string) int {
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
