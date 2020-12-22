package event_handle

import (
	socketio "github.com/googollee/go-socket.io"
	selfUtils "github.com/headend/agent-gateway-service/utils"
	"github.com/headend/share-module/configuration"
	socket_event "github.com/headend/share-module/configuration/socket-event"
	wupModel "github.com/headend/share-module/model/warmup"
	"github.com/headend/share-module/MQ"
	"log"
	"time"
)

const interval = 10
func Warmup(conf configuration.Conf, server *socketio.Server) {
	for {
		var warmupMessageData wupModel.WarmupMessage
		var warmupData []wupModel.WarmupElement
		server.ForEach("/", socket_event.NhomChung, func(conn socketio.Conn) {
			var wupElement wupModel.WarmupElement
			wupElement.IP , wupElement.Port = selfUtils.GetIpAndPortFromRemoteAddr(conn.RemoteAddr().String())
			wupElement.Status = true
			warmupData = append(warmupData, wupElement)
		})
		// Make warup sting
		warmupMessageData.EventTime = time.Now().Unix()
		warmupMessageData.Data = warmupData
		warmupMessageString, err := warmupMessageData.GetJsonString()
		if err != nil {
			log.Println(err)
			time.Sleep(interval * time.Second)
			continue
		}
		// push to message queue
		var mq messagequeue.MQ
		mq.PushMsgByTopic(&conf, warmupMessageString, conf.MQ.WarmUpTopic)
		if mq.Err != nil {
			log.Println(mq.Err )
			time.Sleep(interval * time.Second)
			continue
		}
		time.Sleep(interval * time.Second)
	}
}