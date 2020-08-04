package main

// UserCredentials for body of request of the apis
type userCreds struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type conversationCreds struct {
	RivalUsername string `json:"rivalUsername"`
}

type messageCreds struct {
	Body          string `json:"body"`
	RivalUsername string `json:"rivalUsername"`
}

type changePassCreds struct {
	Password    string `json:"password"`
	NewPassword string `json:"newPassword"`
}

// other objects

type conversationDetails struct {
	ConvID        int    `json:"convID"`
	RivalUsername string `json:"rivalUsername"`
}

type messageReturnType struct {
	Body     string `json:"body"`
	DateTime string `json:"dateTime"`
	Sent     bool   `json:"sent"`
}
