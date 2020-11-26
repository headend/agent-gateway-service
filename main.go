package main

import (
	"github.com/headend/agent-gateway-service/model"
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/googollee/go-socket.io"
	"log"
	"net/http"
	"os"
	"github.com/headend/agent-gateway-service/utils"
)


func main() {
	server := socketio.NewServer(nil)
	server.OnConnect("/", func(s socketio.Conn) error {
		s.SetContext("")
		fmt.Println("connected:", s.ID())
		fmt.Println("Allow connect from ip: ", s.RemoteAddr())
		s.Emit("connection", "Welcome to iptv system for Headend department!")
		server.JoinRoom("/", "monitor", s)
		return nil
	})

	server.OnEvent("/", "notice", func(s socketio.Conn, msg string) {
		fmt.Println("notice:", msg)
		s.Emit("reply", "have "+msg)
	})

	server.OnEvent("/log", "msg", func(s socketio.Conn, msg string) string {
		s.SetContext(msg)
		return "recv " + msg
	})

	server.OnEvent("/", "bye", func(s socketio.Conn) string {
		last := s.Context().(string)
		s.Emit("bye", last)
		s.Close()
		return last
	})

	server.OnError("/", func(s socketio.Conn, e error) {
		fmt.Println("meet error:", e)
	})

	server.OnDisconnect("/", func(s socketio.Conn, reason string) {
		fmt.Println("Disconnect from: ", s.RemoteAddr())
		fmt.Println("closed", reason)
	})

	go server.Serve()
	defer server.Close()

	http.Handle("/socket.io/", server)
	http.Handle("/", http.FileServer(http.Dir("./asset")))
	log.Println("Serving at localhost:8000...")
	reader := bufio.NewReader(os.Stdin)
	go func() {
		for {
			data, _, _ := reader.ReadLine()
			command := string(data)
			/*
			Send message to rom
			 */
			//rom := "monitor"
			//SendMessageToRom(server, rom, command)
			/*
			End: send message to rom
			 */

			/*
			Send file info to rom
			 */
			// get md5 from file
			md5str, err := utils.GetMd5FromFile("asset/" + command)
			if err != nil{
				panic(err)
			} else {
				fileSize, err := utils.GetFileSizeInByte("asset/" + command)
				if err != nil{
					panic(err)
				} else {
					fileInfo2Send := model.WorkerUpdateSignal{
						FileName:       command,
						FilePath:       command,
						FileSizeInByte: fileSize,
						Md5:            md5str,
					}
					b, err := json.Marshal(fileInfo2Send)
					if err != nil {
						fmt.Printf("Error: %s", err)
					}
					server.BroadcastToRoom("/", "monitor", "file", string(b))
					log.Printf("File info to send:%v\n", string(b))
				}
			}
			/*
			End send file info to rom
			 */
		}
	}()
	log.Fatal(http.ListenAndServe(":8000", nil))
}

func SendMessageToRom(server *socketio.Server, rom string, command string) {
	server.BroadcastToRoom("/", rom, "message", command)
	log.Printf("Client connected:%v\n", server.Count())
}
