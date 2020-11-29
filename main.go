package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/googollee/go-socket.io"
	"github.com/headend/share-module/configuration"
	"github.com/headend/share-module/configuration/static-config"
	"github.com/headend/share-module/file-and-directory"
	share_model "github.com/headend/share-module/model"
	"log"
	"net/http"
	"os"
)


func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	// load config
	var conf  configuration.Conf
	conf.LoadConf()

	/*
		Xử lý thông tin kết nối
		Nếu thông tin không có trong config thì lấy từ static config
	*/

	var gwHost string
	if conf.AgentGateway.Gateway != "" {
		gwHost = conf.AgentGateway.Gateway
	} else {
		if conf.AgentGateway.Host != "" {
			gwHost = conf.AgentGateway.Host
		} else {
			gwHost = static_config.GatewayHost
		}
	}


	var gwPort uint16
	if conf.AgentGateway.Port != 0 {
		gwPort = conf.AgentGateway.Port
	} else {
		gwPort = static_config.GatewayPort
	}
	// make socket
	server, err := socketio.NewServer(nil)
	if err != nil {
		log.Fatal(err)
	}
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
	//http.Serve(ln,server)
	http.Handle("/", http.FileServer(http.Dir("./asset")))

	listenAddress := fmt.Sprintf("%s:%d", gwHost, gwPort)
	log.Println("Serving at ", listenAddress)
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
			md5str, err := file_and_directory.GetMd5FromFile("asset/" + command)
			if err != nil{
				panic(err)
			} else {
				fileSize, err := file_and_directory.GetFileSizeInByte("asset/" + command)
				if err != nil{
					panic(err)
				} else {
					fileInfo2Send := share_model.WorkerUpdateSignal{
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
	// runserver here
	log.Fatal(http.ListenAndServe(listenAddress, nil))
}

func SendMessageToRom(server *socketio.Server, rom string, command string) {
	server.BroadcastToRoom("/", rom, "message", command)
	log.Printf("Client connected:%v\n", server.Count())
}
