package main

import (
	"fmt"
	"github.com/googollee/go-socket.io"
	"github.com/headend/agent-gateway-service/event-handle"
	selfUtils "github.com/headend/agent-gateway-service/utils"
	"github.com/headend/share-module/configuration"
	"github.com/headend/share-module/configuration/socket-event"
	"github.com/headend/share-module/model"
	"log"
	"net/http"
)


func main() {

	// load config
	var conf configuration.Conf
	conf.LoadConf()

	/*
		Xử lý thông tin kết nối
		Nếu thông tin không có trong config thì lấy từ static config
	*/

	gwHost, gwPort := selfUtils.GetGWConnectionInfo(conf)
	// make socket
	server, err := socketio.NewServer(nil)
	if err != nil {
		log.Fatal(err)
	}
	server.OnConnect("/", func(s socketio.Conn) error {
		return event_handle.ListenConnection(s, server, &conf)
	})

	server.OnEvent("/", socket_event.ThongBao, func(s socketio.Conn, msg string) {
		event_handle.OnNotice(s, msg)
	})

	server.OnEvent("/log", socket_event.NhanLog, func(s socketio.Conn, msg string) string {
		return event_handle.OnLog(s, msg)
	})
	server.OnEvent("/", socket_event.KetQuaThucThiLenh, func(s socketio.Conn, msg string) {
		content := fmt.Sprintf("On %s result: %s", s.RemoteAddr(), msg)
		log.Print(content)
	})

	server.OnEvent("/", "monitor-response", func(s socketio.Conn, msg string) {
		content := fmt.Sprintf("On %s result: %s", s.RemoteAddr(), msg)
		log.Print(content)
		go func() {
			var onProfileChangeStatus model.ProfileChangeStatus
			err := onProfileChangeStatus.LoadFromJsonString(msg)
			if err != nil {
				log.Println(err)
				return
			}
			if onProfileChangeStatus == (model.ProfileChangeStatus{}) {
				err2 := fmt.Errorf("Invalid data input on profile change status type: %s", msg)
				log.Println(err2)
				return
			}
			// Now update profile
			
		}()
	})

	server.OnEvent("/", "profile-monitor-request", func(s socketio.Conn, msg string) {
		content := fmt.Sprintf("On %s result: %s", s.RemoteAddr(), msg)
		log.Print(content)
		go func() {

			var onProfileChangeStatus model.ProfileChangeStatus
			err := onProfileChangeStatus.LoadFromJsonString(msg)
			if err != nil {
				log.Println(err)
				return
			}
			if onProfileChangeStatus == (model.ProfileChangeStatus{}) {
				err2 := fmt.Errorf("Invalid data input on profile change status type: %s", msg)
				log.Println(err2)
				return
			}
			// Now update profile

		}()
	})

	server.OnEvent("/", "bye", func(s socketio.Conn) string {
		last := s.Context().(string)
		s.Emit("bye", last)
		s.Close()
		return last
	})

	server.OnError("/", func(s socketio.Conn, e error) {
		event_handle.OnErr(s, e)
	})

	server.OnDisconnect("/", func(s socketio.Conn, reason string) {
		event_handle.OnDisconnection(s, reason, &conf)
	})

	go server.Serve()
	defer server.Close()
	http.Handle("/socket.io/", server)
	//http.Serve(ln,server)
	http.Handle("/", http.FileServer(http.Dir("./asset")))

	listenAddress := fmt.Sprintf("%s:%d", gwHost, gwPort)
	log.Println("Serving at ", listenAddress)

	// [AgentCTL]listen message from control
	go event_handle.AgentControl(conf, server)
	// [Warmup] check client connect, then send warmup message
	go event_handle.Warmup(conf, server)
	// runserver here
	log.Fatal(http.ListenAndServe(listenAddress, nil))
}
















