package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/gorilla/websocket"
)

type PackageJSON struct {
	Capacity int    `json:"capacity"`
	ChunkID  int    `json:"chunk_id"`
	Data     string `json:"data"`
}

var countImg = 0
var data bytes.Buffer
var isFirstMsg = true

const ChunkSize = 100000

func makeData(chunk []byte) {
	data.Write(chunk)
}

func makeFile() {
	// log.Println(string(data.Bytes()))
	if err := ioutil.WriteFile(fmt.Sprint("image_", countImg, ".png"), data.Bytes(), 0644); err != nil {
		panic(err)
	}
	countImg++
}

func onMessage(message []byte) {
	// bodyObj := &PackageJSON{}

	// if err := json.Unmarshal(message, bodyObj); err != nil {
	// 	log.Println(err.Error())
	// 	return
	// }

	// log.Println(bodyObj.ChunkID)

	// if bodyObj.ChunkID > 1 && isFirstMsg {
	// 	return
	// } else {
	// 	isFirstMsg = false
	// }

	// data, _ := base64.StdEncoding.DecodeString(bodyObj.Data)

	if len(message) < ChunkSize {
		// log.Println(string(bodyObj.Data))
		makeData(message)
		makeFile()
	}

	makeData(message)
}

func main() {
	cInt, _, err := websocket.DefaultDialer.Dial("ws://127.0.0.1:5959/interactive?key=freeman", nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer cInt.Close()

	c, _, err := websocket.DefaultDialer.Dial("ws://127.0.0.1:5959/stream?key=freeman", nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			return
		}

		log.Println("RECV: ", string(message))

		onMessage(message)
	}
}
