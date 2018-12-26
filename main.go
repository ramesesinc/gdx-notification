package main

import (
	"io/ioutil"
	"log"
	"net/http"

	"github.com/go-redis/redis"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var redisdb = redis.NewClient(&redis.Options{
	Addr:     "localhost:6379",
	Password: "",
	DB:       0,
})

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func monitorHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	id := vars["id"]
	pubsub := redisdb.Subscribe(id)
	_, err = pubsub.Receive()
	if err != nil {
		log.Fatal(err)
		return
	}
	ch := pubsub.Channel()
	for msg := range ch {
		conn.WriteMessage(websocket.TextMessage, []byte(msg.Payload))
		pubsub.Close()
		break
	}
}

func monitorCompletedHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	body, _ := ioutil.ReadAll(r.Body)
	err := redisdb.Publish(id, string(body)).Err()
	if err != nil {
		log.Fatal(err)
		return
	}
}

func main() {
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/gdx-notifier/subcribe/{id}", monitorHandler)
	router.HandleFunc("/gdx-notifier/publish/{id}", monitorCompletedHandler).Methods("POST")
	log.Fatal(http.ListenAndServe(":8082", router))
}
