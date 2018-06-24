package main

import (
	"crypto/tls"
	"io"
	"log"
	"mono/lib/print"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type WebSocketProxy struct {
	serverURL *url.URL
}

func NewWebsocketProxy(serverURL string) *WebSocketProxy {
	url, _ := url.Parse(serverURL)
	return &WebSocketProxy{
		serverURL: url,
	}
}

func (p *WebSocketProxy) handle(w http.ResponseWriter, r *http.Request) {
	if !isWebsocketRequest(r) {
		log.Println("Not a Websocket request")
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	log.Println("Proxying WS connection")
	print.PrintRequest(r)
	var err error
	// Connect to WS server
	var wsConn net.Conn
	if p.serverURL.Scheme == "ws" {
		wsConn, err = net.Dial("tcp", p.serverURL.Host)
	} else {
		wsConn, err = tls.Dial("tcp", p.serverURL.Host, &tls.Config{})
	}
	log.Println("Connecting to", p.serverURL.Host)
	if err != nil {
		log.Println("Error connecting to WS server", p.serverURL, err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	log.Println("Connected")
	defer wsConn.Close()
	// Take over the connection
	hj, ok := w.(http.Hijacker)
	if !ok {
		log.Println("Server doesn't support hijacking")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	conn, _, err := hj.Hijack()
	if err != nil {
		log.Println("Could not hijack the connection", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	conn.SetDeadline(time.Time{}) // Clear deadline
	defer conn.Close()
	// Forward the underlying request to WS server
	wsReq := makeWebsocketRequest(r, p.serverURL)
	err = wsReq.Write(wsConn)
	if err != nil {
		log.Println("Error connecting to WS server", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	log.Println("Proxied WS connection")
	errc := make(chan error, 2)
	go proxy(errc, wsConn, conn)
	go proxy(errc, conn, wsConn)
	err = <-errc
	log.Println("Closing proxied connection")
	if err != nil {
		log.Println("Socket error", err)
	}
}

func proxy(errc chan<- error, dst io.Writer, src io.Reader) {
	// Blocks and copies from src to dst until either EOF is reached on src or
	// an error occurs.
	_, err := io.Copy(dst, src)
	errc <- err
}

func makeWebsocketRequest(r *http.Request, wsServerURL *url.URL) *http.Request {
	wsReq := new(http.Request)
	*wsReq = *r
	wsReq.URL = wsServerURL
	wsReq.Host = wsServerURL.Host
	header := http.Header{}
	// Copy the following headers
	for _, v := range []string{
		"Connection",
		"Upgrade",
		"User-Agent",
		"Origin",
		"Sec-Websocket-Key",
		"Sec-Websocket-Version",
		"Sec-Websocket-Extensions",
		"Cookie",
		"X-Request-Id",
		"X-Request-Start",
	} {
		header.Add(v, r.Header.Get(v))
	}
	wsReq.Header = header
	return wsReq
}

func isWebsocketRequest(req *http.Request) bool {
	connection := strings.ToLower(req.Header.Get("Connection")) == "upgrade"
	upgrade := strings.ToLower(req.Header.Get("Upgrade")) == "websocket"
	return connection && upgrade
}
