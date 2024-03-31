package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	// Uncomment this block to pass the first stage

	listener, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	conn, err := listener.Accept()
	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
		os.Exit(1)
	}
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		fmt.Println("Error reading data: ", err.Error())
		return
	}
	data := string(buf[:n])
	log.Println("Received data", data)

	lines := strings.Split(data, "\r\n")
	firstLinePart := strings.Split(lines[0], " ")
	verb := firstLinePart[0]
	path := firstLinePart[1]
	version := firstLinePart[2]
	log.Println("Verb", verb)
	log.Println("Path", path)
	log.Println("Version", version)
	var response string
	if path == "/" {
		response = "200 OK"
	} else {
		response = "404 Not Found"
	}
	_, err = conn.Write([]byte("HTTP/1.1 " + response + "\r\n\r\n"))
	if err != nil {
		fmt.Println("Error writing data: ", err.Error())
		return
	}
}
