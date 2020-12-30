package event_handle

import (
	"fmt"
	socketio "github.com/googollee/go-socket.io"
	agentpb "github.com/headend/iptv-agent-service/proto"
	messagequeue "github.com/headend/share-module/MQ"
	"github.com/headend/share-module/configuration"
	"github.com/headend/share-module/configuration/socket-event"
	static_config "github.com/headend/share-module/configuration/static-config"
	"github.com/headend/share-module/model/register"
	selfUtils "github.com/headend/agent-gateway-service/utils"
	"github.com/headend/share-module/model"
	"log"
	"strings"
	"time"
)

func ListenConnection(s socketio.Conn, server *socketio.Server, conf *configuration.Conf) error {
	s.SetContext("")
	fmt.Println("connected:", s.ID())
	fmt.Println("Allow connect from ip: ", s.RemoteAddr())
	s.Emit("connection", "Welcome to iptv system for Headend department!")
	server.JoinRoom("/", socket_event.NhomChung, s)
	// make register
	var nraData register.Register
	remoteAddrArr := strings.Split(s.RemoteAddr().String(), ":")
	agentdIp, agentdPort := remoteAddrArr[0], remoteAddrArr[1]
	nraData.IP = agentdIp
	nraData.Port = agentdPort
	nraData.EventTime = time.Now().Unix()
	nraMessageString, err := nraData.GetJsonString()
	if err != nil {
		log.Println(err)
		return err
	} else {
		var queueServer messagequeue.MQ
		queueServer.PushMsgByTopic(conf, nraMessageString, conf.MQ.NraTopic)
	}
	// Sync worker
	AgentInfo := selfUtils.GetAgentByIP(conf.RPC.Agent.Gateway, conf.RPC.Agent.Port, agentdIp)
	if AgentInfo != nil {
		if AgentInfo.IsMonitor {
			if AgentInfo.SignalMonitor {
				err2 := initAgentdWorkerType(s,AgentInfo,static_config.StartMonitorSignal)
				if err2 != nil {
					return err2
				}
			}
			if AgentInfo.VideoMonitor {
				err3 := initAgentdWorkerType(s,AgentInfo,static_config.StartMonitorVideo)
				if err3 != nil {
					return err3
				}
			}
		}
	}
	return nil
}

func initAgentdWorkerType(s socketio.Conn, AgentInfo *agentpb.Agent, ControlType int) error {
	controlData := model.AgentCTLQueueRequest{
		AgentCtlRequest: model.AgentCtlRequest{
			AgentId:     AgentInfo.Id,
			ControlId:   0,
			ControlType: ControlType,
			RunThread:   int(AgentInfo.RunThread),
			TunnelData:  nil,
		},
		ControlType: ControlType,
		EventTime:   time.Now().Unix(),
	}
	ctlMessageString, err := controlData.GetJsonString()
	if err != nil {
		log.Println(err)
		return err
	} else {
		s.Emit(socket_event.DieuKhien, ctlMessageString)
	}
	return nil
}
