package main

import (
	"aziraphale.top/note/cache"
	"encoding/json"
	"golang.org/x/net/websocket"
	"io/ioutil"
	"log"
	"net/http"
)

const NOTE string = "REDIS:NOTE:CONTENT"
const WAKE string = "REDIS:NOTE:CONTENT:CHANGE"

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

	http.Handle("/api/echo", websocket.Handler(noteServer))
	http.HandleFunc("/api/newest", newest)
	http.HandleFunc("/api/update", update)
	if err := http.ListenAndServe(":60443", nil); err != nil {
		log.Fatal(err)
	}
}
