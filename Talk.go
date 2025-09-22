package main

import (
	"bufio"
	"fmt"
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
		log.Println(err)
	}
	defer conn.Close()
	fmt.Println("Client: Connection established.")

	readChan := make(chan []byte)
	writeChan := make(chan []byte)
	quitChan := make(chan bool)
	errChan := make(chan error)

	// Infinite loop until a break command is received
CONNECTION_LOOP:
	for {
		// A go function that reads data from the connection into a buffer
		// Once the data from the connection has been read into the buffer,
		// send the data from the buffer into the read channel.
		go func() {
			buffer := make([]byte, 1024)

			n, err := conn.Read(buffer)
			if err != nil {
				errChan <- err
				return
			}

			readChan <- buffer[:n]
		}()

		// A go function to read input from the user into a buffer.
		// Once data has been read from input, it is sent to the write channel,
		// where it can then be written to the connection in the select statement.
		go func() {
			buffer := make([]byte, 1024)

			reader := bufio.NewReader(os.Stdin)

			n, err := reader.Read(buffer)
			if err != nil {
				errChan <- err
				return
			}

			if strings.EqualFold(string(buffer[:n]), "quit") {
				quitChan <- strings.EqualFold(string(buffer[:n]), "quit")
				return
			}

			writeChan <- buffer[:n]
		}()

		// A select statement that waits until there is something in one of the channels
		// to do something with.
		select {
		case readData := <-readChan:
			fmt.Println("Message Received: ", string(readData))

		case writeData := <-writeChan:
			conn.Write(writeData)

		case errData := <-errChan:
			fmt.Println("Error occurred: ", errData)

		case quitData := <-quitChan:
			fmt.Println("Ending Connection: ", quitData)
			break CONNECTION_LOOP
		}
	}

}

func serverMode() {

	// Establish a tcp server that listens on port 8080, defer closure until the end of the program.
	ln, err := net.Listen("tcp", "localhost:8080")
	if err != nil {
		log.Fatalln(err)
	}
	defer ln.Close()

	go func() {
		//Accept incoming connection
		conn, err := ln.Accept()
		if err != nil {
			log.Println("Error accepting connection: ", err)
			return
		}

		conn.Write([]byte("Server: Connection established."))
	}()
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	//Create a byte buffer of size 1024
	buffer := make([]byte, 1024)

	for {
		//Read from the connection into the byte buffer
		n, err := conn.Read(buffer)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				//The read operation simply timed out, move on
				continue
			} else {
				log.Println(err)
				break
			}
		} else if n > 0 {
			//Check if the int returned by reading the connection is more than 0. If it is, there is something to write.
			//Print what has been read into the byte buffer.
			//The statement buffer[:n] will read everything in the buffer up to but excluding the nth element
			fmt.Println(string(buffer[:n]))
		}
	}
}
