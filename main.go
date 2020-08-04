package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {

	db := dbConn()
	db.Close()

	http.HandleFunc("/users", getUsers)

	http.HandleFunc("/logout", logoutHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/signup", createUser)
	http.HandleFunc("/change-password", changePassword)

	// http.HandleFunc("/create-conv", createConversation)

	http.HandleFunc("/send-message", sendMessage)

	// all the handlers for getting the data only

	http.HandleFunc("/get-all-my-conv", getAllMyConv)
	http.HandleFunc("/get-all-messages", getAllMessages)

	fmt.Println("The server is running in port 8000......")
	log.Fatal(http.ListenAndServe("localhost:8000", nil))
}
