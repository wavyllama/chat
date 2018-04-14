package db

import (
	"bytes"
	"fmt"
	"log"
)

type Conversation struct {
	SSID, sentOrReceived int
	message, timestamp string
}

// Creates the conversations table
func SetupConversationTable() {
	log.Println("Creating the conversations table...")
	var createTableCommand bytes.Buffer
	createTableCommand.WriteString("CREATE TABLE IF NOT EXISTS ")
	createTableCommand.WriteString(conversationTableName)
	createTableCommand.WriteString(" (\n")
	createTableCommand.WriteString("SSID INT NOT NULL, \n")
	createTableCommand.WriteString("message varchar(10000) NOT NULL, \n")
	createTableCommand.WriteString("timestamp varchar(30) NOT NULL, \n")
	createTableCommand.WriteString("sent_or_received TINYINT NOT NULL")
	createTableCommand.WriteString(" );")
	ExecuteDatabaseCommand(createTableCommand.String())
}

func InsertIntoConversations(SSID int, message string, timestamp string, sentOrReceived int) {
	log.Println("Inserting data into conversations...")
	if sentOrReceived != 0 && sentOrReceived != 1 {
		fmt.Printf("Invalid entry for sent/received - msut be 0 or 1. Instead, received a %d", sentOrReceived)
	}
	insertCommand := fmt.Sprintf("INSERT INTO %s VALUES (%d, \"%s\", \"%s\", %d)", conversationTableName, SSID, message, timestamp, sentOrReceived)
	ExecuteDatabaseCommand(insertCommand)
}


func QueryConversations() [] Conversation{
	log.Println("Retrieving data from conversations...")
	query := "SELECT * FROM " + conversationTableName;
	return ExecuteConversationQuery(query)
}

// Executes the specified database command
func ExecuteConversationQuery(query string) [] Conversation {
	results, err := DB.Query(query)
	if err != nil {
		fmt.Printf("Failed to execute query %s: %s", query, err)
		panic(err)
	}
	var conversations [] Conversation
	conv := Conversation{}
	for results.Next(){
		err = results.Scan(&conv.SSID, &conv.message, &conv.timestamp, &conv.sentOrReceived)
		if err!= nil {
			fmt.Printf("Failed to parse results %s: %s", query, err)
			panic(err)
		}
		conversations = append(conversations, conv)
	}
	err = results.Err()
	if err != nil {
		log.Fatal(err)
	}
	return conversations
}
