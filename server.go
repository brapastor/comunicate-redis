package main

import (
	"fmt"
	"gopkg.in/redis.v3"
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"net/http"
)
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Request struct {
	Id int `json:"id"`
	Name string `json:"name"`
}
type Client struct{
	Id int
	websocket *websocket.Conn
}


var Clients = make(map[int]Client)

func ConnectNewClient(request_chanel chan Request)  {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		Password:"",
		DB:0,
	})
	pubsub, err := client.Subscribe("test1")
	if err != nil{
		fmt.Println("No es posible suscribirse al canal")
	}
	for {
		message, err := pubsub.ReceiveMessage()
		if err != nil{
			fmt.Println("No es posible ller el mensaje")
		}
		request := Request{}
		if err := json.Unmarshal([]byte(message.Payload), &request); err != nil{
			fmt.Println("No es posible leer el json")
		}
		fmt.Println(request.Name)
		fmt.Println(request.Id)
		request_chanel <- request
		// fmt.Println(message.Channel)
		// fmt.Println(message.Payload)
	}
	fmt.Println("Todo fue exitoso")
}
func main()  {
	channel_request := make(chan Request)
	go ConnectNewClient(channel_request)
	go ValidateChannel(channel_request)
	muxx := mux.NewRouter()
	muxx.HandleFunc("/subscribe", Subscribe).Methods("GET")
	http.Handle("/", muxx)
	fmt.Println("El servidor se encuentra en el puerto 8000")
	http.ListenAndServe(":8000", nil)
}

func Subscribe(w http.ResponseWriter, r *http.Request){
	//upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	ws, err := upgrader.Upgrade(w,r,nil)

	if err != nil {
		return
	}
	fmt.Println("Nuevo web socket")
	count := len(Clients)
	new_client := Client{count, ws}
	Clients[count] = new_client
	fmt.Println("Nuevo Cliente")

	for{
		_, _, err := ws.ReadMessage()
		if err != nil{
			delete(Clients, new_client.Id)
			fmt.Println("se fue el cliente")
			return
		}
	}
}

func ValidateChannel(request chan Request){
	for {
		select{
			case r := <- request:
			// Enviar mensaje
				SendMessage(r)
		}
	}
}

func SendMessage(request Request){
	for _, client := range Clients{
		if err := client.websocket.WriteJSON(request); err != nil{
			return
		}
	}
}