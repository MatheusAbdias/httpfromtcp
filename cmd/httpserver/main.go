package main

import (
	"crypto/sha256"
	"fmt"
	"httpfromtcp/internal/constants"
	"httpfromtcp/internal/headers"
	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
	"httpfromtcp/internal/server"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
)

const port = 42069

const badRequestResponse = `<html>
  <head>
    <title>400 Bad Request</title>
  </head>
  <body>
    <h1>Bad Request</h1>
    <p>Your request honestly kinda sucked.</p>
  </body>
</html>\n`

const internalServerErrorResponse = `<html>
  <head>
    <title>500 Internal Server Error</title>
  </head>
  <body>
    <h1>Internal Server Error</h1>
    <p>Okay, you know what? This one is on me.</p>
  </body>
</html>\n`

const okResponse = `
<html>
  <head>
    <title>200 OK</title>
  </head>
  <body>
    <h1>Success!</h1>
    <p>Your request was an absolute banger.</p>
  </body>
</html>\n`

const httpBinUrl = "https://httpbin.org"

func main() {
	server, err := server.Serve(port, handler)

	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}

func handler(req *request.Request, resp *response.Writer) {
	if req.RequestLine.RequestTarget == "/video" {
		handleBinary(req, resp)
	}
	if strings.HasPrefix(req.RequestLine.RequestTarget, "/httpbin/") {
		handleProxy(req, resp)
		return
	}

	if req.RequestLine.RequestTarget == "/yourproblem" {
		handler400(req, resp)
		return
	}
	if req.RequestLine.RequestTarget == "/myproblem" {
		handler500(req, resp)
		return
	}

	handler200(req, resp)
}

func handleBinary(_ *request.Request, w *response.Writer) {
	var h headers.Headers
	content, err := os.ReadFile("./assets/vim.mp4")
	if err != nil {
		w.WriteStatusLine(response.InternalServerError)
		h = response.GetDefaultHeaders(0)
	} else {
		w.WriteStatusLine(response.OK)
		h = response.GetDefaultHeaders(len(content))
	}

	h.Override(constants.ContentTypeHeader, "video/mp4")
	w.WriteHeaders(h)
	w.WriteBody(content)
}

func handler400(_ *request.Request, w *response.Writer) {
	w.WriteStatusLine(response.BadRequest)
	body := []byte(badRequestResponse)
	h := response.GetDefaultHeaders(len(body))
	h.Override("Content-Type", "text/html")
	w.WriteHeaders(h)
	w.WriteBody(body)
}

func handler500(_ *request.Request, w *response.Writer) {
	w.WriteStatusLine(response.InternalServerError)
	body := []byte(internalServerErrorResponse)
	h := response.GetDefaultHeaders(len(body))
	h.Override(constants.ContentTypeHeader, "text/html")
	w.WriteHeaders(h)
	w.WriteBody(body)
}

func handler200(_ *request.Request, w *response.Writer) {
	w.WriteStatusLine(response.OK)
	body := []byte(okResponse)
	h := response.GetDefaultHeaders(len(body))
	h.Override(constants.ContentTypeHeader, "text/html")
	w.WriteHeaders(h)
	w.WriteBody(body)
}

func handleProxy(req *request.Request, w *response.Writer) {
	path := strings.TrimPrefix(req.RequestLine.RequestTarget, "/httpbin/")

	w.WriteStatusLine(response.OK)
	h := response.GetDefaultHeaders(0)
	h.Delete(constants.ContentLengthHeader)
	h.Set(constants.TransferEncoding, "chunked")
	h.Set(constants.TrailersHeader, "X-Content-SHA256, X-Content-Length")
	w.WriteHeaders(h)

	url := fmt.Sprintf("%s/%v", httpBinUrl, path)
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
		return
	}

	buffer := make([]byte, 1024)
	var responseBody []byte
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		for {
			n, err := resp.Body.Read(buffer)
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Panic(err)
				return
			}
			_, err = w.WriteChunkedBody(buffer[:n])
			if err != nil {
				log.Panic(err)
				return
			}

			responseBody = append(responseBody, buffer[:n]...)
		}

		_, err = w.WriteChunkedBodyDone()
		if err != nil {
			log.Panic(err)
			return
		}

		bodyCheckSum := sha256.Sum256(responseBody)

		trailers := headers.Headers{
			"X-Content-Sha256": fmt.Sprintf("%x", bodyCheckSum),
			"X-Content-Length": strconv.Itoa(len(responseBody)),
		}

		err := w.WriteTrailers(trailers)
		if err != nil {
			log.Panic(err)
			return
		}
	}
}
