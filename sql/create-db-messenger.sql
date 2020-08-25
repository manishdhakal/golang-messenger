CREATE DATABASE messenger;
USE messenger;

CREATE TABLE users (
	id INT NOT NULL PRIMARY KEY AUTO_INCREMENT,
    username VARCHAR(50) NOT NULL UNIQUE,
	password_hash VARCHAR(64) NOT NULL -- sha256 encoding for password hashing
);

CREATE TABLE conversations (
	id INT NOT NULL PRIMARY KEY AUTO_INCREMENT,
    user1_id INT NOT NULL,
	user2_id INT NOT NULL,
    FOREIGN KEY (user1_id) REFERENCES users(id),
    FOREIGN KEY (user2_id) REFERENCES users(id)
);

CREATE TABLE messages (
	id INT NOT NULL PRIMARY KEY AUTO_INCREMENT,
    body VARCHAR(5000) NOT NULL,
	conversation_id INT NOT NULL,
    sender_id INT NOT NULL,
    reciever_id INT NOT NULL,
    date_time DATETIME NOT NULL,
    FOREIGN KEY (conversation_id) REFERENCES conversations(id),
    FOREIGN KEY (sender_id) REFERENCES users(id),
    FOREIGN KEY (reciever_id) REFERENCES users(id)
);

CREATE TABLE sessions (
	session_id VARCHAR(64) NOT NULL UNIQUE,  -- sha256 encoding for session hashingconversations
    user_id INT NOT NULL,
    expiry DATE NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id)
);