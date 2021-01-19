package main

import (
	"fmt"
	"github.com/googollee/go-socket.io"
	"github.com/headend/agent-gateway-service/event-handle"
	selfUtils "github.com/headend/agent-gateway-service/utils"
	"github.com/headend/share-module/configuration"
	"github.com/headend/share-module/configuration/socket-event"
	static_config "github.com/headend/share-module/configuration/static-config"
	file_and_directory "github.com/headend/share-module/file-and-directory"
	agentModel "github.com/headend/share-module/model/agentd"
	shareModel "github.com/headend/share-module/model"
	"log"
	"net/http"
	"os"
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

	//server.OnEvent("/", socket_event.ThongBao, func(s socketio.Conn, msg string) {
	//	event_handle.OnNotice(s, msg)
	//})

	//server.OnEvent("/log", socket_event.NhanLog, func(s socketio.Conn, msg string) string {
	//	return event_handle.OnLog(s, msg)
	//})
	//server.OnEvent("/", socket_event.KetQuaThucThiLenh, func(s socketio.Conn, msg string) {
	//	content := fmt.Sprintf("On %s result: %s", s.RemoteAddr(), msg)
	//	log.Print(content)
	//})

	server.OnEvent("/", socket_event.MonitorResponse, func(s socketio.Conn, msg string) {
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
			monitorInfo, err2 := selfUtils.OnMonitorChangeUpdateStatus(conf, onProfileChangeStatus)
			if err2 != nil {
				log.Println(err)
			}
			// do write log
			logData := selfUtils.MakeLogInDataRequest(onProfileChangeStatus, monitorInfo)
			selfUtils.DoWriNonitorLog(conf, logData)
		}()
	})

	server.OnEvent("/", socket_event.ProfileRequest, func(s socketio.Conn, monitorType string) {
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
					s.Emit(socket_event.ProfileResponse, ProfileDataToAgentListString)
				}
			}()
		}
	})

	server.OnEvent("/", socket_event.PongPong, func(s socketio.Conn, pongMsg string) {
		if cliWorkerVer, err := strconv.ParseFloat(pongMsg, 32); err == nil {
			// get worker version on server
			worker_version_file := static_config.WorkerVersionFile
			filee := file_and_directory.MyFile{Path: worker_version_file}
			newWorkerVerStr, err1 := filee.Read()
			if err1 != nil {
				log.Println(err)
			} else {
				if newWorkerVer, err2 := strconv.ParseFloat(newWorkerVerStr, 32); err2 == nil {
					if newWorkerVer != cliWorkerVer {
						// send to control update message
						var updateMessage string
						controlData := shareModel.AgentCtlRequest{
							AgentId:     0,
							ControlId:   0,
							ControlType: static_config.UpdateWorker,
							RunThread:   0,
							TunnelData:  nil,
						}
						updateMessage, _ = controlData.GetJsonString()
						s.Emit(socket_event.DieuKhien, updateMessage)
						// send to update handle
						filee := file_and_directory.MyFile{Path: static_config.WorkerVersionFile}
						var updateInfo shareModel.WorkerUpdateSignal
						updateInfo.FilePath = static_config.WorkerVersionFile
						updateInfo.FileName = static_config.AgentdWorkerName
						updateInfo.Md5, _ = filee.GetMd5FromFile(static_config.WorkerVersionFile)
						updateInfo.FileSizeInByte, _ = filee.GetFileSizeInByte(static_config.WorkerVersionFile)
						updateInfoStr, _ := updateInfo.GetJsonString()
						s.Emit(socket_event.UpdateWorker, updateInfoStr)
					}
				} else {
					log.Println(err)
				}
			}
		} else {
			log.Println(err)
		}
	})

	server.OnEvent("/", "worker-update-request", func(s socketio.Conn, Msg string) {
		filee := file_and_directory.MyFile{Path: static_config.WorkerVersionFile}
		var updateInfo shareModel.WorkerUpdateSignal
		updateInfo.FilePath = static_config.WorkerVersionFile
		updateInfo.FileName = static_config.AgentdWorkerName
		updateInfo.Md5, _ = filee.GetMd5FromFile(static_config.WorkerVersionFile)
		updateInfo.FileSizeInByte, _ = filee.GetFileSizeInByte(static_config.WorkerVersionFile)
		updateInfoStr, _ := updateInfo.GetJsonString()
		s.Emit(socket_event.UpdateWorker, updateInfoStr)
	})

	server.OnEvent("/", "sync-worker", func(s socketio.Conn, Msg string) {
		agentdIp, _ := selfUtils.GetIpAndPortFromRemoteAddr(s.RemoteAddr().String())
		// Sync worker
		AgentInfo := selfUtils.GetAgentByIP(conf.RPC.Agent.Gateway, conf.RPC.Agent.Port, agentdIp)
		if AgentInfo != nil {
			if AgentInfo.IsMonitor {
				if AgentInfo.SignalMonitor {
					err2 := selfUtils.InitAgentdWorkerType(s, AgentInfo, static_config.StartMonitorSignal)
					if err2 != nil {
						log.Println(err2)
					}
				}
				if AgentInfo.VideoMonitor {
					err3 := selfUtils.InitAgentdWorkerType(s, AgentInfo, static_config.StartMonitorVideo)
					if err3 != nil {
						log.Println(err3)
					}
				}
				if AgentInfo.AudioMonitor {
					err4 := selfUtils.InitAgentdWorkerType(s, AgentInfo, static_config.StartMonitorAudio)
					if err4 != nil {
						log.Println(err4)
					}
				}
			}
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

	// create storage folder
	if _, err7 := os.Stat(static_config.GatewayStorage); os.IsNotExist(err7) {
		err8 := os.Mkdir(static_config.GatewayStorage, 0755)
		panic(err8)
	}

	go server.Serve()
	defer server.Close()
	http.Handle("/socket.io/", server)
	//http.Serve(ln,server)

	http.Handle("/", http.FileServer(http.Dir(static_config.GatewayStorage)))

	listenAddress := fmt.Sprintf("%s:%d", gwHost, gwPort)
	log.Println("Serving at ", listenAddress)

	// [AgentCTL]listen message from control
	go event_handle.AgentControl(conf, server)
	// [Warmup] check client connect, then send warmup message
	go event_handle.Warmup(conf, server)
	// runserver here
	log.Fatal(http.ListenAndServe(listenAddress, nil))
}





















