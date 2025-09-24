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

	handleComm(conn, "Client")

	/*
			//Cutoff for what starts to be the communications handling.
			readChan := make(chan []byte)
			writeChan := make(chan []byte)
			quitChan := make(chan bool)
			errChan := make(chan error)

			// A mutex var so that we can lock while the read function is working
			var muRead sync.Mutex

			// A mutex var so that we can lock while the write function is working
			var muWrite sync.Mutex

			// Infinitely loop until a break command is received
		COMM_LOOP:
			for {
				// A go function that reads data from the connection into a buffer
				// Once the data from the connection has been read into the buffer,
				// send the data from the buffer into the read channel.
				go func() {
					muRead.Lock()
					defer muRead.Unlock()
					fmt.Println("We have entered reading mode")
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
					muWrite.Lock()
					defer muWrite.Unlock()
					fmt.Println("We have entered writing mode")
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
					fmt.Println(string(readData))

				case writeData := <-writeChan:
					conn.Write(writeData)

				case errData := <-errChan:
					/*
						if netErr, ok := errData.(net.Error); ok && netErr.Timeout() {
							log.Println("Timeout Error: ", errData)
							continue COMM_LOOP
						} else {
							fmt.Println("Error occurred: ", errData)
							log.Fatalln(errData)
						}*/ /*

			fmt.Println("Error occurred: ", errData)
			log.Fatalln(errData)

		case quitData := <-quitChan:
			fmt.Println("Ending Connection: ", quitData)
			break COMM_LOOP

		default:
		}
	}*/
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

	handleComm(conn, "Server")
}

// A function for handling all communication over a connection, both reading and writing.
func handleComm(conn net.Conn, modeName string) {
	defer conn.Close()
	ack := (modeName + ": Connection Established")
	conn.Write([]byte(ack))

	readChan := make(chan []byte)
	writeChan := make(chan []byte)
	quitChan := make(chan bool)
	errChan := make(chan error)

	// A mutex var so that we can lock while the read function is working
	//var muRead sync.Mutex

	// A mutex var so that we can lock while the write function is working
	//var muWrite sync.Mutex

	// A go function that reads data from the connection into a buffer
	// Once the data from the connection has been read into the buffer,
	// send the data from the buffer into the read channel.

	/*
		go func() {
			muRead.Lock()
			defer muRead.Unlock()
			fmt.Println("We have entered reading mode")
			buffer := make([]byte, 1024)

			n, err := conn.Read(buffer)
			if err != nil {
				errChan <- err
				return
			}

			readChan <- buffer[:n]
		}()*/

	/*
		// A go function to read input from the user into a buffer.
		// Once data has been read from input, it is sent to the write channel,
		// where it can then be written to the connection in the select statement.
		go func() {
			muWrite.Lock()
			defer muWrite.Unlock()
			fmt.Println("We have entered writing mode")
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
		}()*/

	go handleRead(conn, readChan, errChan)
	go handleWrite(conn, writeChan, errChan, quitChan)

	// Infinite loop over the select statement to do polling
COMM_LOOP:
	for {
		// A select statement that waits until there is something in one of the channels to do something with.
		select {
		case readData := <-readChan:
			fmt.Println(string(readData))
			go handleRead(conn, readChan, errChan)

		case writeData := <-writeChan:
			conn.Write(writeData)
			go handleWrite(conn, writeChan, errChan, quitChan)

		case errData := <-errChan:
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
	//muRead.Lock()
	//defer muRead.Unlock()
	fmt.Println("We have entered reading mode")
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
// Takes a net.Conn connection, a []byte channel to send the buffer into, an error channel to send any errors
// into for later error handling, and a bool channel to send input that is flagged to end the connection into.
func handleWrite(conn net.Conn, wch chan []byte, ech chan error, bch chan bool) {
	//muWrite.Lock()
	//defer muWrite.Unlock()
	fmt.Println("We have entered writing mode")
	buffer := make([]byte, 1024)

	reader := bufio.NewReader(os.Stdin)

	n, err := reader.Read(buffer)
	if err != nil {
		ech <- err
		return
	}

	if strings.EqualFold(string(buffer[:n]), "quit") {
		bch <- strings.EqualFold(string(buffer[:n]), "quit")
		return
	}

	wch <- buffer[:n]
}
