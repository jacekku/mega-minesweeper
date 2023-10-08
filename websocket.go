package main

import (
	"encoding/binary"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var whole_request = time.Now()

const whole_map_header byte = 'm'
const uncover_header byte = 'u'

func update_board(clients *[]*websocket.Conn, fields chan []idx_val) {
	for {
		select {
		case vals := <-fields:
			start := time.Now()
			var msg []byte = make([]byte, 0)
			msg = append(msg, uncover_header)
			for _, i_v := range vals {
				msg = binary.LittleEndian.AppendUint64(msg, uint64(i_v.idx))
				msg = binary.LittleEndian.AppendUint64(msg, uint64(i_v.val))
			}
			// fmt.Println(msg)
			// fmt.Println(len(*clients))
			fmt.Println(time.Since(start), "create")
			broadcast(*clients, websocket.BinaryMessage, msg)
			fmt.Println(time.Since(start), "create and broadcast")
			fmt.Println(time.Since(whole_request), "whole request")
		}
	}
}

func broadcast(clients []*websocket.Conn, msgType int, msg []byte) {
	start := time.Now()
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
	fmt.Println(time.Since(start), "broadcast")
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

func main() {
	var clients []*websocket.Conn = make([]*websocket.Conn, 0)

	var brd = create_board(10, 15)
	fmt.Println("generated board")
	for i := 0; i < int(float64(brd.height*brd.width)*0.1); i++ {
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
		fmt.Println("connected")
		if err = client.WriteMessage(websocket.TextMessage, fow.byte_metadata()); err != nil {
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
		// if err = client.WriteMessage(websocket.BinaryMessage, []byte{0xffff}); err != nil {
		// return
		// }
		clients = append(clients, client)
		for {
			// Read message from browser
			_, msg, err := client.ReadMessage()
			if err != nil {
				return
			}

			// Print the message to the console
			// fmt.Printf("%s sent: %s\n", client.RemoteAddr(), string(msg))

			// Write message back to browser
			whole_request = time.Now()
			start := time.Now()
			if msg[0] == 'u' {
				var s_idx = string(msg[1:])
				var idx, err = strconv.ParseUint(s_idx, 10, 32)
				if err != nil {
					fmt.Println(err)
					continue
				}
				uncoverWithChannel(uint(idx), &fow, updated_fields)
			} else if msg[0] == 'm' {
				var s_idx = string(msg[1:])
				var idx, err = strconv.ParseUint(s_idx, 10, 32)
				if err != nil {
					fmt.Println(err)
					continue
				}
				markWithChannel(uint(idx), &fow, updated_fields)
			}
			fmt.Println(time.Since(start), "handle client command")
		}
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "websockets.html")
	})

	fmt.Println("listening on port 8080")
	http.ListenAndServe(":8080", nil)
}

// func main() {
// 	var brd = create_board(4, 4)
// 	for i := 0; i < 3; i++ {
// 		brd.set_bomb(uint(rand.Intn(int(brd.width))), uint(rand.Intn(int(brd.height))))
// 	}
// 	brd.calculate_foo()
// 	var fow = create_fow(brd)
// 	for _, row := range fow.pretty_print() {
// 		fmt.Println(row)
// 	}

// var i, j uint
// for {
// 	fmt.Print("Type two numbers: ")
// 	fmt.Scan(&i, &j)
// 	fmt.Println("Your numbers are:", i, "and", j)
// 	fmt.Println(fow.uncover(i, j))
// 	for _, row := range fow.pretty_print() {
// 		fmt.Println(row)
// 	}
// }
// }
