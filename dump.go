package main

// // to create conversation
// func createConversation(w http.ResponseWriter, r *http.Request) {

// 	sessionCookie, sessionErr := r.Cookie("session_id")
// 	if sessionErr != nil {
// 		w.WriteHeader(http.StatusUnauthorized)
// 		return
// 	}
// 	sessionID := sessionCookie.Value

// 	var conv conversation
// 	jsonErr := json.NewDecoder(r.Body).Decode(&conv)
// 	if jsonErr != nil {
// 		// If the structure of the body is wrong, return an HTTP error
// 		w.WriteHeader(http.StatusBadRequest)
// 		return
// 	}

// 	db := dbConn()
// 	defer db.Close()

// 	authUserID := getAuthID(sessionID)

// 	if authUserID == 0 {
// 		w.WriteHeader(http.StatusUnauthorized)
// 		return
// 	}

// 	rivalUserID := getUserID(conv.RivalUsername)
// 	if rivalUserID == 0 {
// 		w.WriteHeader(http.StatusBadRequest)
// 		return
// 	}
// 	qStr := fmt.Sprintf(`SELECT EXISTS (
// 			 	SELECT * FROM conversations WHERE (user1_id = %d AND user2_id = %d) OR (user1_id = %d AND user2_id = %d)
// 			) existence;`, authUserID, rivalUserID, rivalUserID, authUserID)
// 	existRow, existErr := db.Query(qStr)
// 	if existErr != nil {
// 		w.WriteHeader(http.StatusInternalServerError)
// 		fmt.Print("Error to check the validation of already existing conv\n")
// 		return
// 	}
// 	var exists bool
// 	existRow.Next()
// 	existRow.Scan(&exists)

// 	if exists {
// 		w.WriteHeader(http.StatusConflict)
// 		fmt.Fprint(w, "The conversation already exists\n")
// 		return
// 	}

// 	qStr = fmt.Sprintf(`INSERT INTO conversations (user1_id, user2_id)
// 				VALUES (%d, %d)`, authUserID, rivalUserID)
// 	_, InsErr := db.Query(qStr)
// 	if InsErr != nil {
// 		w.WriteHeader(http.StatusInternalServerError)
// 	}

// 	fmt.Print("The conversation has been created.\n")

// }
