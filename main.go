package main

import (
	"aziraphale.top/note/cache"
	"crypto/tls"
	"crypto/x509"
	"embed"
	"encoding/json"
	"golang.org/x/net/websocket"
	"io/fs"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

const NOTE string = "REDIS:NOTE:CONTENT"
const WAKE string = "REDIS:NOTE:CONTENT:CHANGE"

//go:embed ssl/ca.crt
var ca []byte

//go:embed ssl/server.crt
var cert []byte

//go:embed ssl/server.key
var key []byte

//go:embed web/build
var html embed.FS

type Response struct {
	Status uint32      `json:"status"`
	Data   interface{} `json:"data"`
}

func noteServer(ws *websocket.Conn) {

	defer ws.Close()
	cookie, err := ws.Request().Cookie("sessionId")
	if err != nil || cookie.Value == "" {
		ws.Close()
		return
	}
	sessionId := cookie.Value
	log.Printf("session of %s connect", sessionId)

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

func update(writer http.ResponseWriter, request *http.Request) {
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
		Addr:         ":60443",
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		TLSConfig: &tls.Config{
			ServerName:   "note.sheffery.cloud",
			ClientAuth:   tls.RequireAndVerifyClientCert,
			ClientCAs:    clientCertPool,
			Certificates: []tls.Certificate{serverCert},
		},
	}

	root, err := fs.Sub(html, "web/build")
	if err != nil {
		panic("static file error")
	}
	http.Handle("/", http.FileServer(http.FS(root)))
	http.Handle("/api/echo", websocket.Handler(noteServer))
	http.HandleFunc("/api/newest", newest)
	http.HandleFunc("/api/update", update)
	if err := server.ListenAndServeTLS("", ""); err != nil {
		log.Fatal(err)
	}
}
