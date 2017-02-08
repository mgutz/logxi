// Copyright 2015 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"text/template"
	"time"

	"github.com/gorilla/websocket"
)

var (
	addr      = flag.String("addr", "127.0.0.1:8080", "http service address")
	cmdPath   string
	homeTempl *template.Template
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Maximum message size allowed from peer.
	maxMessageSize = 8192

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Time to wait before force close on connection.
	closeGracePeriod = 10 * time.Second
)

// func pumpStdin(ws *websocket.Conn, w io.Writer) {
// 	defer ws.Close()
// 	ws.SetReadLimit(maxMessageSize)
// 	ws.SetReadDeadline(time.Now().Add(pongWait))
// 	ws.SetPongHandler(func(string) error { ws.SetReadDeadline(time.Now().Add(pongWait)); return nil })
// 	for {
// 		_, message, err := ws.ReadMessage()
// 		if err != nil {
// 			break
// 		}
// 		message = append(message, '\n')
// 		if _, err := w.Write(message); err != nil {
// 			break
// 		}
// 	}
// }

func pumpStdin(ws *websocket.Conn, r io.Reader, done chan struct{}) {
	defer func() {
	}()
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		ws.SetWriteDeadline(time.Now().Add(writeWait))

		line := scanner.Bytes()

		var m jsonObject

		err := json.Unmarshal(line, &m)
		if err != nil {
			b, _ := json.Marshal(jsonObject{"stdin": string(line)})
			line = b
		}

		if err = ws.WriteMessage(websocket.TextMessage, line); err != nil {
			ws.Close()
			break
		}
	}

	if scanner.Err() != nil {
		log.Println("scan:", scanner.Err())
	}
	close(done)

	ws.SetWriteDeadline(time.Now().Add(writeWait))
	ws.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	time.Sleep(closeGracePeriod)
	ws.Close()
}

func ping(ws *websocket.Conn, done chan struct{}) {
	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if err := ws.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(writeWait)); err != nil {
				log.Println("ping:", err)
			}
		case <-done:
			return
		}
	}
}

func internalError(ws *websocket.Conn, msg string, err error) {
	log.Println(msg, err)
	ws.WriteMessage(websocket.TextMessage, []byte("Internal server error."))
}

var upgrader = websocket.Upgrader{}

func serveWs(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("upgrade:", err)
		return
	}

	defer ws.Close()
	stdoutDone := make(chan struct{})
	go pumpStdin(ws, os.Stdin, stdoutDone)
	go ping(ws, stdoutDone)

	select {
	case <-stdoutDone:
	}
}

func serveHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, "Not found", 404)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	homeTempl.Execute(w, r.Host)
}

func main() {
	// Go 1.8
	progname, err := os.Executable()
	if err != nil {
		panic("Could not read executable path")
	}

	templateFile := filepath.Join(filepath.Dir(progname), "home.html")

	homeTempl = template.Must(template.ParseFiles(templateFile))
	http.HandleFunc("/", serveHome)
	http.HandleFunc("/ws", serveWs)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
