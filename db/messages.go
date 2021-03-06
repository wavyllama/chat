package db

import (
	"bytes"
	"database/sql"
	"github.com/wavyllama/chat/config"
	"time"
	"crypto/aes"
	"crypto/cipher"
)

const (
	Sent     = 0
	Received = 1
)

// Stores a DB Message
type Message struct {
	SSID           uint64
	SentOrReceived int
	Text           []byte
	Timestamp      time.Time
}

// Inserts a message into the messages table
func InsertMessage(SSID uint64, message []byte, timestamp string, sentOrReceived int) bool {
	if sentOrReceived != Sent && sentOrReceived != Received {
		Logger.Fatalf("Invalid entry for sent/received - must be 0 or 1. Instead, received a %d", sentOrReceived)
	}
	encryptedMessages := AESEncrypt(message, []byte(db.AESKey))
	insertCommand, err := DB.Prepare("INSERT INTO messages VALUES (?, ?, ?, ?)")
	if err != nil {
		Logger.Printf("Error creating messages prepared statement for InsertMessage: %s", err)
	}
	_, err = insertCommand.Exec(SSID, encryptedMessages, timestamp, sentOrReceived)
	if err != nil {
		Logger.Panicf("Failed to insert into messages: %s", err)
	}
	return true
}

// Removes a message from the messages table
func DeleteMessage(SSID uint64, message []byte, timestamp string, sentOrReceived int) bool {
	deleteCommand, err := DB.Prepare("DELETE FROM messages WHERE SSID=? AND message=? AND message_timestamp=? AND sent_or_received=?")
	if err != nil {
		Logger.Fatalf("Error creating users prepared statement for UpdateUserIP: %s", err)
	}
	enc := AESEncrypt(message, []byte(db.AESKey))
	_, err = deleteCommand.Exec(SSID, enc, timestamp, sentOrReceived)
	if err != nil {
		Logger.Printf("Failed to delete message: %s", err)
	}
	return true
}

// Returns all data in the messages table
func QueryMessages() []Message {
	query, err := DB.Prepare("SELECT * FROM messages")
	if err != nil {
		Logger.Fatalf("Error creating messages prepared statement for QueryMessages: %s", err)
	}
	results, err := query.Query()
	if err != nil {
		Logger.Printf("Error executing QueryMessages query: %s", err)
	}
	return ExecuteMessagesQuery(results)
}

// Returns all messages for a given session identified by SSID
func getSessionMessages(SSID uint64) []Message {
	query, err := DB.Prepare("SELECT * FROM messages WHERE SSID=?")
	if err != nil {
		Logger.Fatalf("Error creating messages prepared statement for getSessionMessages: %s", err)
	}
	results, err := query.Query(SSID)
	if err != nil {
		Logger.Printf("Error executing getSessionMessages query: %s", err)
	}
	return ExecuteMessagesQuery(results)
}

// Executes the specified database command
func ExecuteMessagesQuery(results *sql.Rows) []Message {
	var messages []Message
	msg := Message{}
	for results.Next() {
		var timestamp string

		err := results.Scan(&msg.SSID, &msg.Text, &timestamp, &msg.SentOrReceived)
		if err != nil {
			Logger.Panicf("Failed to parse results from messages: %s", err)
			panic(err)
		}
		parsedTime, _ := time.Parse("2006-01-02 15:04:05", timestamp)
		msg.Timestamp = parsedTime
		msg.Text = []byte(AESDecrypt(msg.Text, []byte(db.AESKey)))
		messages = append(messages, msg)
	}
	err := results.Err()
	if err != nil {
		Logger.Panicf("Failed to get results from conversations query: %s", err)
	}
	return messages
}

// Parts of AES encryption code from https://gist.github.com/saoin/b306746041b48a8366d0f63507a4e7f3
func AESEncrypt(content []byte, key []byte) []byte {
	block, err := aes.NewCipher(key)
	if err != nil {
		Logger.Fatalf("Error with key: %s", err.Error())
	}
	ecb := cipher.NewCBCEncrypter(block, []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
	content = pkcs5Padding(content, block.BlockSize())
	crypted := make([]byte, len(content))
	ecb.CryptBlocks(crypted, content)
	return crypted
}

func AESDecrypt(crypt []byte, key []byte) []byte {
	block, err := aes.NewCipher(key)
	if err != nil {
		Logger.Fatalf("Error with key: %s", err.Error())
	}
	if len(crypt) == 0 {
		Logger.Println("plain content empty")
	}
	ecb := cipher.NewCBCDecrypter(block, []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
	decrypted := make([]byte, len(crypt))
	ecb.CryptBlocks(decrypted, crypt)
	return pkcs5Trimming(decrypted)
}

func pkcs5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext) % blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func pkcs5Trimming(encrypt []byte) []byte {
	padding := encrypt[len(encrypt)-1]
	return encrypt[:len(encrypt)-int(padding)]
}
