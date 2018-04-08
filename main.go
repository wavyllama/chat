package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"github.com/chat/server"
)

const Exit = "exit"

// Listen to standard in for messages to be sent
func listen(program *server.Server) {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		message := scanner.Text()
		if message == Exit {
			return
		}
		stringSlice := strings.Fields(message)
		if err := program.Send(stringSlice[0], strings.Join(stringSlice[1:], " ")); err != nil {
			fmt.Printf("input: %s\n", err.Error())
		}
	}
	if scanner.Err() != nil {
		fmt.Printf(scanner.Err().Error())
		return
	}
}

func main() {
	var program server.Server
	if err := program.Start("Archil"); err != nil {
		log.Fatalf("main: %s", err.Error())
	}
	defer program.Shutdown()
	sig := make (chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-sig
		os.Exit(0)
	}()
	listen(&program)
}