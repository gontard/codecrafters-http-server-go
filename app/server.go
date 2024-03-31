package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
)

func main() {
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
	CRLF := "\r\n"
	if path == "/" {
		_, err = conn.Write([]byte("HTTP/1.1 200 OK" + CRLF + CRLF))
	} else if strings.HasPrefix(path, "/echo/") {
		value := strings.TrimPrefix(path, "/echo/")
		_, err = conn.Write([]byte("HTTP/1.1 200 OK" + CRLF +
			"Content-Type: text/plain" + CRLF +
			"Content-Length: " + strconv.Itoa(len(value)) + CRLF + CRLF +
			value + CRLF))
	} else {
		_, err = conn.Write([]byte("HTTP/1.1 404 Not Found" + CRLF + CRLF))
	}
	if err != nil {
		fmt.Println("Error writing data: ", err.Error())
		return
	}
}
