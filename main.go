package main

import (
	"crypto/tls"
	"crypto/x509"
	"embed"
	"encoding/json"
	"io/fs"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"aziraphale.top/note/cache"
	"golang.org/x/net/websocket"
)

const NOTE string = "REDIS:NOTE:CONTENT"
const WAKE string = "REDIS:NOTE:CONTENT:CHANGE"

//go:embed ssl_dev/ca.crt
var ca []byte

//go:embed ssl_dev/server.crt
var cert []byte

//go:embed ssl_dev/server.key
var key []byte

//go:embed static
var html embed.FS

// Echo the data received on the WebSocket.
func noteServer(ws *websocket.Conn) {

	defer ws.Close()
	cookie, err := ws.Request().Cookie("sessionId")
	if err != nil || cookie.Value == "" {
		ws.Close()
		return
	}
	sessionId := cookie.Value
	log.Printf("session of %s connect", sessionId)

	// init content
	ws.Write([]byte(cache.Get(NOTE)))

	bell := cache.Listen(WAKE)
	for {
		latest := <-bell
		if latest != sessionId {
			_, err := ws.Write([]byte(cache.Get(NOTE)))
			if err != nil {
				break
			}
		}
	}
	log.Printf("session of %s lost connection", sessionId)
}

// This example demonstrates a trivial echo server.
func main() {

	clientCertPool := x509.NewCertPool()
	ok := clientCertPool.AppendCertsFromPEM(ca)
	if !ok {
		panic("failed to parse root certificate")
	}

	serverCert, err := tls.X509KeyPair(cert, key)
	if err != nil {
		panic("failed to parse server certificate")
	}

	server := &http.Server{
		Addr:         ":443",
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		TLSConfig: &tls.Config{
			ServerName:   "sync.note.cloud",
			ClientAuth:   tls.RequireAndVerifyClientCert,
			ClientCAs:    clientCertPool,
			Certificates: []tls.Certificate{serverCert},
		},
	}

	root, err := fs.Sub(html, "static")
	if err != nil {
		panic("static file error")
	}
	http.Handle("/", http.FileServer(http.FS(root)))
	http.Handle("/echo", websocket.Handler(noteServer))
	http.HandleFunc("/newest", newest)
	http.HandleFunc("/update", updatefunc)
	if err := server.ListenAndServeTLS("", ""); err != nil {
		log.Fatal(err)
	}
}

func updatefunc(writer http.ResponseWriter, request *http.Request) {
	cookie, err := request.Cookie("sessionId")
	if err != nil || cookie.Value == "" {
		var response Response = Response{
			Status: 400,
		}
		responseByte, _ := json.Marshal(response)
		writer.Write(responseByte)
		return
	}
	content, _ := ioutil.ReadAll(request.Body)
	cache.SetWithNoExpire(NOTE, string(content))
	cache.Bell(WAKE, cookie.Value)
	var response Response = Response{
		Status: 200,
	}
	responseByte, _ := json.Marshal(response)
	writer.Write(responseByte)
}

func newest(writer http.ResponseWriter, request *http.Request) {
	content := cache.Get(NOTE)
	var response Response = Response{
		Status: 200,
		Data:   content,
	}
	responseByte, _ := json.Marshal(response)
	writer.Write(responseByte)
}

type Response struct {
	Status uint32      `json:"status"`
	Data   interface{} `json:"data"`
}
