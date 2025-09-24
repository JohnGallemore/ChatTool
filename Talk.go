package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
)

func main() {

	//Verify that arguments were given, then verify that those arguments were valid. If so, launch in the specified mode.
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "-h":
			fmt.Println("Application launched in Client Mode")
			clientMode()

		case "-s":
			fmt.Println("Application launched in Server Mode")
			serverMode()

		case "-a":
			fmt.Println("Application launched in Auto Mode")

		case "-help":
			fmt.Println("Application launched in Help Mode")

		default:
			fmt.Println("Your arguments were invalid, relaunch using -help argument.")
		}
	}
}

func clientMode() {

	// Establish a connection by trying to connect to a provided IP address.
	// Currently only using localhost:8080 for testing
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		log.Println("Error creating connection: ", err)
		return
	}

	name := []byte("Client: ")

	// Call handleComm to handle communcations over this connection.
	handleComm(conn, name)
}

func serverMode() {

	// Establish a tcp server that listens on port 8080, defer closure until the end of the program.
	ln, err := net.Listen("tcp", "localhost:8080")
	if err != nil {
		log.Fatalln(err)
	}
	defer ln.Close()

	//Accept incoming connection
	conn, err := ln.Accept()
	if err != nil {
		fmt.Println("Error accepting connection: ", err)
		return
	}

	name := []byte("Server: ")

	// Call handleComm to handle communications over this connection.
	handleComm(conn, name)
}

// A function for handling all communication over a connection, both reading and writing.
// Takes a net.Conn connection and a []byte slice representing the name of the user that calls handleComm
func handleComm(conn net.Conn, name []byte) {
	defer conn.Close()
	ack := (string(name) + "Connection Established\n")
	conn.Write([]byte(ack))

	readChan := make(chan []byte)
	writeChan := make(chan []byte)
	quitChan := make(chan bool)
	errChan := make(chan error)

	go handleRead(conn, readChan, errChan)
	go handleWrite(writeChan, errChan, quitChan)

	// Infinite loop over the select statement to do polling
COMM_LOOP:
	for {
		// A select statement that waits until there is something in one of the channels to do something with.
		select {
		case readData := <-readChan:
			fmt.Printf("%s\n", string(readData))
			go handleRead(conn, readChan, errChan)

		case writeData := <-writeChan:
			msg := append(name, writeData...)
			conn.Write(msg)
			go handleWrite(writeChan, errChan, quitChan)

		case errData := <-errChan:
			if errors.Is(errData, io.EOF) {
				fmt.Println("Termination: Other participant has disconnected.")
				break COMM_LOOP
			}
			fmt.Println("Error occurred: ", errData)
			log.Fatalln(errData)

		case quitData := <-quitChan:
			fmt.Println("Ending Connection: ", quitData)
			break COMM_LOOP

		}
	}
}

// A function that reads data from the connection into a buffer
// Once the data from the connection has been read into the buffer,
// send the data from the buffer into the read channel.
// Takes a net.Conn connection, a []byte channel to send the buffer into, and an error channel to send any errors
// into for later error handling.
func handleRead(conn net.Conn, rch chan []byte, ech chan error) {
	buffer := make([]byte, 1024)

	n, err := conn.Read(buffer)
	if err != nil {
		ech <- err
		return
	}

	rch <- buffer[:n]
}

// A go function to read input from the user into a buffer.
// Once data has been read from input, it is sent to the write channel,
// where it can then be written to the connection in the select statement.
// Since this handles user input on the user side, it does not need to know anything about the connection,
// this function only needs to know which channels to send information to.
// Takes a []byte channel to send the buffer into, an error channel to send any errors
// into for later error handling, and a bool channel to send input that is flagged to end the connection into.
func handleWrite(wch chan []byte, ech chan error, bch chan bool) {
	buffer := make([]byte, 1024)

	reader := bufio.NewReader(os.Stdin)
	fmt.Println()

	n, err := reader.Read(buffer)
	if err != nil {
		ech <- err
		return
	}

	msg := string(buffer[:(n - 2)])

	if strings.EqualFold(msg, "quit") {
		bch <- strings.EqualFold(msg, "quit")
		return
	}

	wch <- buffer[:n]
}
