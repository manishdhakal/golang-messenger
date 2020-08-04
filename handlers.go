package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

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

	var creds userCreds

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

	var access bool
	rows.Next()
	rows.Scan(&access)

	if !access {
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
	http.SetCookie(w, &http.Cookie{
		Name:  "session_id",
		Value: "",
	})

	sessionCookie, sessionErr := r.Cookie("session_id")
	if sessionErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	sessionID := sessionCookie.Value

	// set cookie to be empty
	db := dbConn()
	defer db.Close()
	qStr := fmt.Sprintf(`DELETE FROM sessions WHERE session_id = "%s";`, sessionID)
	_, err := db.Query(qStr)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

}

// to create a user
func createUser(w http.ResponseWriter, r *http.Request) {
	var user userCreds
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
	var exst bool
	exist.Next()
	exist.Scan(&exst)

	if exst {
		w.WriteHeader(http.StatusConflict)
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

func sendMessage(w http.ResponseWriter, r *http.Request) {

	authUserID := getAuthID(w, r)
	var msg messageCreds
	// Get the JSON body and decode into credentials
	jsonErr := json.NewDecoder(r.Body).Decode(&msg)
	if jsonErr != nil {
		// If the structure of the body is wrong, return an HTTP error
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	rivalUserID := getUserID(msg.RivalUsername)

	if authUserID == 0 {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if rivalUserID == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	db := dbConn()
	defer db.Close()
	qStr := fmt.Sprintf(`SELECT id FROM conversations 
			WHERE (user1_id = %d AND user2_id = %d) OR (user1_id = %d AND user2_id = %d);`,
		authUserID, rivalUserID, rivalUserID, authUserID)
	idRow, idErr := db.Query(qStr)

	if idErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Print("Error to check the validation of already existing conv\n")
		return
	}

	var convID int
	if idRow.Next() {
		idRow.Scan(&convID)
	} else {
		qStr = fmt.Sprintf(`INSERT INTO conversations (user1_id, user2_id)
		VALUES (%d, %d)`, authUserID, rivalUserID)
		_, insErr := db.Query(qStr)
		if insErr != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
		fmt.Print("The conversation has been created.\n")

		qStr = fmt.Sprintf(`SELECT id FROM conversations 
		WHERE (user1_id = %d AND user2_id = %d)`, authUserID, rivalUserID)

		idRow, idErr := db.Query(qStr)

		if idErr != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Print("Error to check the validation of already existing conv\n")
			return
		}

		idRow.Next()
		idRow.Scan(&convID)
	}

	qStr = fmt.Sprintf(`INSERT INTO messages (body, conversation_id, sender_id, reciever_id, date_time)
		VALUES ("%s", %d, %d, %d, "%s");`, msg.Body, convID, authUserID, rivalUserID, time.Now().Format("2006-01-02 15:04:05"))

	_, insERR := db.Query(qStr)

	if insERR != nil {
		w.WriteHeader(http.StatusInternalServerError)
		println(insERR.Error())
		return
	}

	fmt.Println("The message has been added.")

}

func getAllMyConv(w http.ResponseWriter, r *http.Request) {
	authUserID := getAuthID(w, r)
	if authUserID == 0 {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	db := dbConn()
	defer db.Close()
	qStr := fmt.Sprintf(`SELECT id, user1_id, user2_id FROM conversations 
		WHERE user1_id = %d OR user2_id = %d;`, authUserID, authUserID)
	convRows, convErr := db.Query(qStr)
	if convErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var convs []conversationDetails
	var convID, user1ID, user2ID int
	var rivalUsername string

	for convRows.Next() {
		convRows.Scan(&convID, &user1ID, &user2ID)
		var rivalUserID int
		if user1ID == authUserID {
			rivalUserID = user2ID
		} else {
			rivalUserID = user1ID
		}
		rivalUsername = getUserName(rivalUserID)
		if rivalUsername == "" {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		convs = append(convs, conversationDetails{ConvID: convID, RivalUsername: rivalUsername})
	}

	json.NewEncoder(w).Encode(convs)

}

func getAllMessages(w http.ResponseWriter, r *http.Request) {

	authUserID := getAuthID(w, r)
	if authUserID == 0 {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	convID, convErr := strconv.Atoi(r.URL.Query().Get("convID"))
	if convErr != nil {
		fmt.Println(convErr)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	db := dbConn()
	defer db.Close()

	qStr := fmt.Sprintf(`SELECT EXISTS (
			SELECT * FROM conversations WHERE id = %d AND (user1_id = %d OR user2_id = %d)
		) existence;`, convID, authUserID, authUserID)

	existsRow, existsErr := db.Query(qStr)
	if existsErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	var exists bool
	existsRow.Next()
	existsRow.Scan(&exists)

	if exists == false {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	qStr = fmt.Sprintf(`SELECt body, sender_id, date_time FROM messages 
		WHERE conversation_id = %d ORDER BY date_time DESC;`, convID)
	msgRows, msgErr := db.Query(qStr)
	if msgErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	var (
		body        string
		senderID    int
		dateTime    string
		messagesRet []messageReturnType
	)

	for msgRows.Next() {
		msgRows.Scan(&body, &senderID, &dateTime)
		var sent bool
		if senderID == authUserID {
			sent = true
		} else {
			sent = false
		}
		messagesRet = append(messagesRet, messageReturnType{Body: body, Sent: sent, DateTime: dateTime})
	}
	json.NewEncoder(w).Encode(messagesRet)

}

func changePassword(w http.ResponseWriter, r *http.Request) {
	authUserID := getAuthID(w, r)
	if authUserID == 0 {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	var creds changePassCreds

	// Get the JSON body and decode into credentials
	jsonErr := json.NewDecoder(r.Body).Decode(&creds)
	fmt.Println(creds, authUserID)
	if jsonErr != nil {
		// If the structure of the body is wrong, return an HTTP error
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	db := dbConn()
	defer db.Close()

	qStr := fmt.Sprintf(`SELECT EXISTS (
		SELECT * FROM users WHERE id = "%d" AND password_hash = SHA2("%s",256)
	  ) access`, authUserID, creds.Password)

	existsRow, existsErr := db.Query(qStr)

	if existsErr != nil {
		fmt.Println(existsErr.Error())

		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	var exists bool
	existsRow.Next()
	existsRow.Scan(&exists)
	if !exists {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	qStr = fmt.Sprintf(`UPDATE users SET password_hash = sha2("%s", 256) WHERE id = %d;`, creds.NewPassword, authUserID)
	_, chngErr := db.Query(qStr)
	if chngErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
