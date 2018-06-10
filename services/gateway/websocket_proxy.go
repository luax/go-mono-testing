package main

import (
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
)

type WebSocketProxy struct {
	target *url.URL
}

func NewWebsocketProxy(target string) *WebSocketProxy {
	url, _ := url.Parse(target)
	return &WebSocketProxy{
		target: url,
	}
}

func (p *WebSocketProxy) handle(w http.ResponseWriter, r *http.Request) {
	if !isWebsocketRequest(r) {
		log.Println("Not a Websocket request")
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	w.Header().Set("X-API-KEY", "VerySecret")
	log.Println("Proxying WS connection")

	// Hijack to take over the connection
	hj, ok := w.(http.Hijacker)
	if !ok {
		log.Fatal("Server doesn't support hijacking")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	conn, _, err := hj.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	// Connect to backend
	backendConn, err := net.Dial("tcp", p.target.Host)
	if err != nil {
		log.Printf("Error connecting to backend %s: %v", p.target, err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer backendConn.Close()

	err = r.Write(backendConn)
	if err != nil {
		log.Printf("Error copying request to target: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	log.Println("Proxied WS connection")

	errc := make(chan error, 2)
	go proxy(errc, backendConn, conn)
	go proxy(errc, conn, backendConn)
	err = <-errc
	log.Println("Closing proxied connection", err)

	if err != nil {
		log.Printf("Socket error: %v", err)
	}
}

func proxy(errc chan<- error, dst io.Writer, src io.Reader) {
	// Blocks and copies from src to dst until either EOF is reached on src or
	// an error occurs.
	_, err := io.Copy(dst, src)
	errc <- err
}

func isWebsocketRequest(req *http.Request) bool {
	connection := strings.ToLower(req.Header.Get("Connection")) == "upgrade"
	upgrade := strings.ToLower(req.Header.Get("Upgrade")) == "websocket"
	return connection && upgrade
}
