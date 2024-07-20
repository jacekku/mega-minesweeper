package main

import (
	"encoding/binary"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

const whole_map_header byte = 'm'
const uncover_header byte = 'u'
const position_header byte = 'p'

func update_board(clients *[]*websocket.Conn, fields chan []idx_val) {
	for {
		select {
		case vals := <-fields:
			var msg []byte = make([]byte, 0)
			msg = append(msg, uncover_header)
			for _, i_v := range vals {
				msg = binary.LittleEndian.AppendUint64(msg, uint64(i_v.idx))
				msg = binary.LittleEndian.AppendUint64(msg, uint64(i_v.val))
			}
			broadcast(*clients, websocket.BinaryMessage, msg)
		}
	}
}

func broadcast(clients []*websocket.Conn, msgType int, msg []byte) {
	for idx := len(clients) - 1; idx >= 0; idx-- {
		c := clients[idx]
		if err := c.WriteMessage(msgType, msg); err != nil {
			fmt.Println(err)
			clients[idx] = clients[len(clients)-1]
			clients = clients[:len(clients)-1]
			c.Close()
			continue
		}
	}
}

func uncoverWithChannel(idx uint, fow *fog_of_war, updated_fields chan []idx_val) {
	var values, _ = fow.uncover(idx)
	updated_fields <- values
}

func markWithChannel(idx uint, fow *fog_of_war, updated_fields chan []idx_val) {
	var value, err = fow.mark(idx)
	if err != nil {
		return
	}
	updated_fields <- []idx_val{value}
}

func updatePositions(clients []*websocket.Conn, idx uint) {
	var msg []byte = make([]byte, 0)
	msg = append(msg, position_header)
	msg = binary.LittleEndian.AppendUint64(msg, uint64(idx))

	broadcast(clients, websocket.BinaryMessage, msg)
}

func main() {

	var width, err1 = strconv.ParseUint(os.Getenv("MS_WIDTH"), 10, 32)
	if err1 != nil {
		fmt.Println("MS_WIDTH", err1)
		width = 200
	}
	var height, err2 = strconv.ParseUint(os.Getenv("MS_HEIGHT"), 10, 32)
	if err2 != nil {
		fmt.Println("MS_HEIGHT", err2)
		height = 200
	}
	var percentage, err3 = strconv.ParseFloat(os.Getenv("MS_PERCENTAGE"), 32)
	if err3 != nil {
		fmt.Println("MS_PERCENTAGE", err3)
		percentage = 0.1
	}
	var port = os.Getenv("MS_PORT")
	if port == "" {
		port = "8080"
	}

	var clients []*websocket.Conn = make([]*websocket.Conn, 0)

	var brd = create_board(uint(width), uint(height))
	fmt.Println("generated board", width, height)
	for i := 0; i < int(float64(brd.height*brd.width)*percentage); i++ {
		brd.set_bomb(uint(rand.Intn(int(brd.width))), uint(rand.Intn(int(brd.height))))
	}
	fmt.Println("Place bombs")
	brd.calculate_foo()
	var updated_fields = make(chan []idx_val)
	var fow = create_fow(brd)
	go update_board(&clients, updated_fields)

	http.HandleFunc("/echo", func(w http.ResponseWriter, r *http.Request) {
		client, err := upgrader.Upgrade(w, r, nil) // error ignored for sake of simplicity
		if err != nil {
			fmt.Errorf("upsi dupsi")
			fmt.Println(err)
			return
		}
		fmt.Println("connected", time.Now(), client.RemoteAddr().String())
		if err = client.WriteMessage(websocket.TextMessage, fow.byte_metadata()); err != nil {
			fmt.Println("could not write message")
			return
		}
		var a = fow.byte_array()
		var msg = make([]byte, len(a)+1)
		msg[0] = whole_map_header
		for i, v := range a {
			msg[i+1] = v
		}
		if err = client.WriteMessage(websocket.BinaryMessage, msg); err != nil {
			return
		}
		clients = append(clients, client)
		for {
			// Read message from browser
			_, msg, err := client.ReadMessage()
			if err != nil {
				return
			}

			// Print the message to the console

			// Write message back to browser
			if msg[0] == 'u' {
				// uncover
				var s_idx = string(msg[1:])
				var idx, err = strconv.ParseUint(s_idx, 10, 32)
				if err != nil {
					fmt.Println(err)
					continue
				}
				uncoverWithChannel(uint(idx), &fow, updated_fields)
			} else if msg[0] == 'm' {
				// mark
				var s_idx = string(msg[1:])
				var idx, err = strconv.ParseUint(s_idx, 10, 32)
				if err != nil {
					fmt.Println(err)
					continue
				}
				markWithChannel(uint(idx), &fow, updated_fields)
			} else if msg[0] == 'p' {
				// update position
				var s_idx = string(msg[1:])
				var idx, err = strconv.ParseUint(s_idx, 10, 32)
				if err != nil {
					fmt.Println(err)
					continue
				}
				updatePositions(clients, uint(idx))
			}
		}
	})

	http.HandleFunc("/assets/", func(w http.ResponseWriter, r *http.Request) {
		var fileName = r.URL.Path
		fileName = strings.Split(fileName, "/")[2]
		http.ServeFile(w, r, "./assets/"+fileName)
	})
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("serving websockets.html")
		http.ServeFile(w, r, "websockets.html")
	})

	http.HandleFunc("/admin/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(r.Body)
		fmt.Println(r.Method)
	})

	fmt.Println("listening on port " + port)
	var serveError = http.ListenAndServe(":"+port, nil)
	if serveError != nil {
		fmt.Println(serveError)
	}
}
