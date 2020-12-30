package main

import (
	"fmt"
	"github.com/googollee/go-socket.io"
	"github.com/headend/agent-gateway-service/event-handle"
	selfUtils "github.com/headend/agent-gateway-service/utils"
	"github.com/headend/share-module/configuration"
	"github.com/headend/share-module/configuration/socket-event"
	agentModel "github.com/headend/share-module/model/agentd"
	"log"
	"net/http"
	"strconv"
)


func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

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
		log.Printf(content)
		go func() {
			var onProfileChangeStatus agentModel.ProfileChangeStatus
			err := onProfileChangeStatus.LoadFromJsonString(msg)
			if err != nil {
				log.Println(err)
				return
			}
			if onProfileChangeStatus == (agentModel.ProfileChangeStatus{}) {
				err2 := fmt.Errorf("Invalid data input on profile change status type: %s", msg)
				log.Println(err2)
				return
			}
			log.Printf("%#v", onProfileChangeStatus)
			// Now update profile
			err2 := selfUtils.OnMonitorChangeUpdateStatus(conf, onProfileChangeStatus)
			if err2 != nil {
				log.Println(err)
			}
		}()
	})

	server.OnEvent("/", "profile-monitor-request", func(s socketio.Conn, monitorType string) {
		content := fmt.Sprintf("On %s result: %s", s.RemoteAddr(), monitorType)
		if mType, err := strconv.Atoi(monitorType); err != nil {
			fmt.Println(content, "is not an integer.")
		} else {
			go func() {
				ip, _ := selfUtils.GetIpAndPortFromRemoteAddr(s.RemoteAddr().String())
				res, err := selfUtils.GetProfileMonitorByAgent(conf, ip, mType)
				if err != nil {
					log.Println(err)
					return
				}
				var AgentdMonitorDataInput agentModel.MonitorInputForAgent
				var ProfileDataToAgentList []agentModel.ProfileForAgentdElement
				for _, profile := range res.Profiles {
					//log.Printf("origin data %#v", profile)
					ProfileDataToAgent := agentModel.ProfileForAgentdElement{
						MonitorId:   profile.MonitorId,
						ProfileId:   profile.MonitorId,
						AgentId:     profile.AgentId,
						Status:      profile.StatusId,
						VideoStatus: profile.StatusVideo,
						MulticastIP: profile.MulticastIp,
					}
					ProfileDataToAgentList = append(ProfileDataToAgentList, ProfileDataToAgent)
				}
				AgentdMonitorDataInput.MonitorType = mType
				AgentdMonitorDataInput.ProfileList = ProfileDataToAgentList

				ProfileDataToAgentListString, err := AgentdMonitorDataInput.GetJsonString()
				log.Println(ProfileDataToAgentListString)
				if err != nil {
					log.Println(err)
				} else {
					s.Emit("profile-monitor-response", ProfileDataToAgentListString)
				}
			}()
		}
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




















