package main

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

var clients = make(map[*websocket.Conn]bool) //접속된 클라이언트
var broadcast = make(chan Message)           //메세지 브로드캐스트

// 업그레이드
var upgrade = websocket.Upgrader{}

// 메세지 JSON 구조체
type Message struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Message  string `json:"message"`
}

func main() {
	// 파일 서버 가동
	fs := http.FileServer(http.Dir("./public"))
	http.Handle("/", fs)
	// websocker으로 루팅을 연결
	http.HandleFunc("/ws", handleConnections)
	go handleMessages()
	log.Println("http server started on :8000")
	err := http.ListenAndServe(":8000", nil)

	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	// 받은 GET 요청 websocket으로 업그레이드
	ws, err := upgrade.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}

	// 함수가 끝나면 소켓을 꼭 닫아야한다.
	defer ws.Close()

	//클라이언트를 새롭게 등록
	clients[ws] = true

	for {
		var msg Message
		// 새로운 메세지를 JSON으로 읽고, Message 오브젝트에 매핑.
		err := ws.ReadJSON(&msg)
		if err != nil {
			log.Printf("error : %v", err)
			delete(clients, ws)
			break
		}
		// 새롭게 수신된 메시지를 브로드케스트 채널에 보낸다.
		broadcast <- msg
	}
}

func handleMessages() {
	for {
		// 브로드 케스트 채널에서 다음 메시지를 받는다.
		msg := <-broadcast
		// 현재 접속중인 모든 클라이언트들에게 메세지를 보낸다.
		for client := range clients {
			err := client.WriteJSON(msg)
			if err != nil {
				log.Printf("error : %v", err)
				client.Close()
				delete(clients, client)
			}
		}
	}
}
