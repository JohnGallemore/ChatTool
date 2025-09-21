package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
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

	// Establish a connection by tryign to connect to a provided IP address.
	// Currently only using localhost:8080 for testing
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		log.Println(err)
	}
	defer conn.Close()

	reader := bufio.NewReader(os.Stdin)

	for {
		msg, err := reader.ReadString('\n')
		if err != nil {
			log.Println(err)
		}

		conn.Write([]byte(msg))
	}
}

func serverMode() {

	// Establish a tcp server that listens on port 8080, defer closure until the end of the program.
	ln, err := net.Listen("tcp", "localhost:8080")
	if err != nil {
		log.Fatalln(err)
	}
	defer ln.Close()

	for {
		//Accept incoming connection
		conn, err := ln.Accept()
		if err != nil {
			log.Println()
		}

		go handleConnection(conn)
	}
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
		} else if n > 0 { //Check if the int returned by reading the connection is more than 0. If it is, there is something to write.
			//Print what has been read into the byte buffer.
			//The statement buffer[:n] will read everything in the buffer up to but excluding the nth element
			fmt.Println(string(buffer[:n]))
		}
	}
}
