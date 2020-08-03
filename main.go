package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {

	db := dbConn()
	db.Close()
	// t := time.Now()

	// fmt.Printf("%v", t.Format("2006-01-02 "))

	http.HandleFunc("/users", getUsers)

	http.HandleFunc("/logout", logoutHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/signup", createUser)

	// http.HandleFunc("/create-conv", createConversation)

	http.HandleFunc("/send-message", sendMessage)

	// all the handlers for getting the data only

	// http.HandleFunc("/get-all-my-conv", getAllMyConv)

	fmt.Println("The server is running in port 8000......")
	log.Fatal(http.ListenAndServe("localhost:8000", nil))
}
