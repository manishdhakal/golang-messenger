package main

// UserCredentials for body of request
type userCredentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type conversation struct {
	RivalUsername string `json:"rivalUsername"`
}

type message struct {
	Body          string `json:"body"`
	RivalUsername string `json:"rivalUsername"`
}
