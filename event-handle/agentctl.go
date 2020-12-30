package event_handle

import (
	"encoding/json"
	"fmt"
	socketio "github.com/googollee/go-socket.io"
	selfUtils "github.com/headend/agent-gateway-service/utils"
	messagequeue "github.com/headend/share-module/MQ"
	"github.com/headend/share-module/configuration"
	socket_event "github.com/headend/share-module/configuration/socket-event"
	"github.com/headend/share-module/model"
	"log"
	"time"
)

func AgentControl(conf configuration.Conf, server *socketio.Server) {
	var mq messagequeue.MQ
	mq.InitConsumerByTopic(&conf, conf.MQ.OperationTopic)
	defer mq.CloseConsumer()
	if mq.Err != nil {
		log.Print(mq.Err)
	}
	log.Printf("Listen mesage from %s topic\n", conf.MQ.OperationTopic)
	for {
		msg, err := mq.Consumer.ReadMessage(-1)
		if err != nil {
			log.Printf("Consumer error to topic %s: %v (%v)\n", conf.MQ.OperationTopic, err, msg)
			log.Printf("Wait for retry %d(s)...", 10)
			time.Sleep(10 * time.Second)
			continue
		}
		log.Println("Connect to message queue successfully!")
		log.Println(string(msg.Value))
		go func() {
			var controlMsgData *model.AgentCTLQueueRequest
			json.Unmarshal(msg.Value, &controlMsgData)
			reportJitter := fmt.Sprintf("Jitter %d", time.Now().Unix()-controlMsgData.EventTime)
			log.Println(reportJitter)
			// get agent info
			//	connect agent services
			var agentid = controlMsgData.AgentId
			var agentServerHost = conf.RPC.Agent.Gateway
			var agentServerPort = conf.RPC.Agent.Port
			res := selfUtils.GetAgentByID(agentServerHost, agentServerPort, err, agentid)
			if err != nil {
				log.Println(err)
				return
			}
			if len(res.Agents) == 0 {
				log.Println("Agent not found")
				return
			}
			// send control message to agent
			//broadcast_to_agentd.SendMessageToRom(server, socket_event.NhomChung, socket_event.DieuKhien, string(msg.Value))
			// Only send message to agent on info
			server.ForEach("/", socket_event.NhomChung, func(conn socketio.Conn) {
				tmpip, _ := selfUtils.GetIpAndPortFromRemoteAddr(conn.RemoteAddr().String())
				if tmpip == res.Agents[0].IpControl {
					conn.Emit(socket_event.DieuKhien, string(msg.Value))
				}
			})
		}()
	}
}
