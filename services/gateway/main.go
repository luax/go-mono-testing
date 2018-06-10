package main

import (
	"html/template"
	"log"
	"mono/lib/env"
	"net/http"
	"os"
)

func handler(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("templates/index.html")
	t.Execute(w, "Hello!")
}

func main() {
	log.Println("Gateway started")
	env.PrintEnvVariables()
	wsProxy := NewWebsocketProxy(os.Getenv("WS_URL"))
	fs := http.FileServer(http.Dir("./static"))
	// HTTP
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	http.HandleFunc("/ws", wsProxy.handle)
	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe(":"+os.Getenv("PORT"), nil))
}
