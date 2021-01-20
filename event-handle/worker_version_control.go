package event_handle

import (
	"fmt"
	socketio "github.com/googollee/go-socket.io"
	"github.com/headend/share-module/configuration"
	socket_event "github.com/headend/share-module/configuration/socket-event"
	static_config "github.com/headend/share-module/configuration/static-config"
	file_and_directory "github.com/headend/share-module/file-and-directory"
	"github.com/headend/share-module/shellout"
	"log"
	"strconv"
	"time"
)

const pingInterval = 60
func PingPing(conf configuration.Conf, server *socketio.Server) {
	/*
	1. cập nhật file version worker
	2. duyệt qua từng agentd và gửi ping message
	*/

	for {
		// 1. cập nhật version worker
		appToRun := fmt.Sprintf("%s/%s", static_config.GatewayStorage, static_config.AgentdWorkerName)
		err, exitCode, stdout, stderr := shellout.RunExternalCmd(appToRun, []string{"-v", "version"}, 5)
		if err != nil{
			log.Println("Can not get worker version from storage")
			log.Println(err)
			time.Sleep(pingInterval * time.Second)
			continue
		}
		if exitCode == 0 {
			if stdout == "" {
				log.Println("Can not get worker version")
				time.Sleep(pingInterval * time.Second)
				continue
			} else {
				if _, errLoadV := strconv.ParseFloat(stdout, 32); err == nil {
					filee := file_and_directory.MyFile{Path: static_config.WorkerVersionFile}
					filee.WriteString(stdout)
				}  else {
					log.Println(errLoadV)
				}
			}
		} else {
			log.Printf("%s - %s", stdout, stderr)
		}

		// 2. gửi ping message
		server.ForEach("/", socket_event.NhomChung, func(conn socketio.Conn) {
			conn.Emit("ping", "ping")
		})
		time.Sleep(pingInterval * time.Second)
	}
}

