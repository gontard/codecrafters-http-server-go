package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

func main() {
	var dir string
	flag.StringVar(&dir, "directory", "", "Directory path to use")
	flag.Parse()
	httpServer := NewHttpServer("0.0.0.0:4221")
	httpServer.AddHandler(echoHandler)
	httpServer.AddHandler(userAgentHandler)
	httpServer.AddHandler(getFileHandler(dir))
	httpServer.AddHandler(postFileHandler(dir))
	httpServer.AddHandler(rootHandler)
	httpServer.AddHandler(notFoundHandler)
	httpServer.ListenAndServe()
}

type HttpReq struct {
	Verb    string
	Path    string
	Version string
	Headers map[string]string
	Body    *string
}

type HttpResp struct {
	Version string
	Status  string
	Headers map[string]string
	Body    *string
}

type HttpServer struct {
	address  string
	Handlers []HttpHandler
}

type HttpHandler func(req *HttpReq) *HttpResp

const CRLF = "\r\n"

func NewHttpServer(address string) *HttpServer {
	return &HttpServer{
		address:  address,
		Handlers: []HttpHandler{},
	}
}

func (s *HttpServer) AddHandler(handler HttpHandler) {
	s.Handlers = append(s.Handlers, handler)
}

func (s *HttpServer) ListenAndServe() {
	listener, err := net.Listen("tcp", s.address)
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		go s.handleClient(conn)
	}
}

func (s *HttpServer) handleClient(conn net.Conn) {
	defer conn.Close()
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		fmt.Println("Error reading data: ", err.Error())
		return
	}
	data := string(buf[:n])
	req := parseRequest(data)
	var resp *HttpResp
	for _, handler := range s.Handlers {
		resp = handler(req)
		if resp != nil {
			break
		}
	}
	_, err = conn.Write([]byte(responseToString(resp)))
	if err != nil {
		fmt.Println("Error writing data: ", err.Error())
		return
	}
}

func echoHandler(req *HttpReq) *HttpResp {
	if strings.HasPrefix(req.Path, "/echo/") {
		headers := map[string]string{}
		body := strings.TrimPrefix(req.Path, "/echo/")
		headers["Content-Type"] = "text/plain"
		headers["Content-Length"] = strconv.Itoa(len(body))
		return newHttpResp("200 OK", headers, &body)
	}
	return nil
}

func userAgentHandler(req *HttpReq) *HttpResp {
	if req.Path == "/user-agent" {
		headers := map[string]string{}
		body := req.Headers["User-Agent"]
		headers["Content-Type"] = "text/plain"
		headers["Content-Length"] = strconv.Itoa(len(body))
		return newHttpResp("200 OK", headers, &body)
	}
	return nil
}

func getFileHandler(dir string) func(req *HttpReq) *HttpResp {
	return func(req *HttpReq) *HttpResp {
		if strings.HasPrefix(req.Path, "/files/") && req.Verb == "GET" {
			fileName := strings.TrimPrefix(req.Path, "/files/")
			fileContent, err := os.ReadFile(dir + "/" + fileName)
			if err == nil {
				body := string(fileContent)
				headers := map[string]string{}
				headers["Content-Type"] = "application/octet-stream"
				headers["Content-Length"] = strconv.Itoa(len(body))
				return newHttpResp("200 OK", headers, &body)
			}
		}
		return nil
	}
}

func postFileHandler(dir string) func(req *HttpReq) *HttpResp {
	return func(req *HttpReq) *HttpResp {
		if strings.HasPrefix(req.Path, "/files/") && req.Verb == "POST" {
			fileName := strings.TrimPrefix(req.Path, "/files/")
			err := os.WriteFile(dir+"/"+fileName, []byte(*req.Body), 0644)
			if err == nil {
				return created()
			}
		}
		return nil
	}
}

func rootHandler(req *HttpReq) *HttpResp {
	if req.Path == "/" {
		return ok()
	}
	return nil
}

func notFoundHandler(_ *HttpReq) *HttpResp {
	return notFound()
}

func parseRequest(data string) *HttpReq {
	lines := strings.Split(data, CRLF)
	firstLinePart := strings.Split(lines[0], " ")
	verb := firstLinePart[0]
	path := firstLinePart[1]
	version := firstLinePart[2]
	headers := make(map[string]string)
	var body string
	for i := 1; i < len(lines); i++ {
		if lines[i] == "" {
			if i+1 < len(lines) {
				body = lines[i+1]
			}
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
		Body:    &body,
	}
}

func responseToString(resp *HttpResp) string {
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

func newHttpResp(status string, headers map[string]string, body *string) *HttpResp {
	version := "HTTP/1.1"
	return &HttpResp{
		Version: version,
		Status:  status,
		Headers: headers,
		Body:    body,
	}
}

func newHttpRespWithStatus(status string) *HttpResp {
	headers := map[string]string{}
	return newHttpResp(status, headers, nil)
}

func ok() *HttpResp {
	return newHttpRespWithStatus("200 OK")
}

func created() *HttpResp {
	return newHttpRespWithStatus("201 Created")
}

func notFound() *HttpResp {
	return newHttpRespWithStatus("404 Not Found")
}
