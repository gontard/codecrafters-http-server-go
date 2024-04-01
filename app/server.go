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
	req := parseRequest(data)
	status := "200 OK"
	headers := map[string]string{}
	var body string
	if strings.HasPrefix(req.Path, "/echo/") {
		body = strings.TrimPrefix(req.Path, "/echo/")
		headers["Content-Type"] = "text/plain"
		headers["Content-Length"] = strconv.Itoa(len(body))
	} else if req.Path == "/user-agent" {
		body = req.Headers["User-Agent"]
		headers["Content-Type"] = "text/plain"
		headers["Content-Length"] = strconv.Itoa(len(body))
	} else if req.Path != "/" {
		status = "404 Not Found"
	}
	resp := &HttpResp{
		Version: "HTTP/1.1",
		Status:  status,
		Headers: headers,
		Body:    &body,
	}
	_, err = conn.Write([]byte(responseToString(resp)))
	if err != nil {
		fmt.Println("Error writing data: ", err.Error())
		return
	}
}

type HttpReq struct {
	Verb    string
	Path    string
	Version string
	Headers map[string]string
}

type HttpResp struct {
	Version string
	Status  string
	Headers map[string]string
	Body    *string
}

func parseRequest(data string) *HttpReq {
	lines := strings.Split(data, "\r\n")
	firstLinePart := strings.Split(lines[0], " ")
	verb := firstLinePart[0]
	path := firstLinePart[1]
	version := firstLinePart[2]
	headers := make(map[string]string)
	for i := 1; i < len(lines); i++ {
		if lines[i] == "" {
			break
		}
		parts := strings.Split(lines[i], ": ")
		headers[parts[0]] = parts[1]
	}
	return &HttpReq{
		Verb:    verb,
		Path:    path,
		Version: version,
		Headers: headers,
	}
}

func responseToString(resp *HttpResp) string {
	CRLF := "\r\n"
	result := resp.Version + " " + resp.Status + CRLF
	for key, value := range resp.Headers {
		result += key + ": " + value + CRLF
	}
	result += CRLF
	if resp.Body != nil {
		result += *resp.Body
	}
	return result
}
