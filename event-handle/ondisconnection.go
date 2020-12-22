package event_handle

import (
	"fmt"
	socketio "github.com/googollee/go-socket.io"
	messagequeue "github.com/headend/share-module/MQ"
	"github.com/headend/share-module/configuration"
	"log"
	"strings"
	"github.com/headend/share-module/model/warmup"
	"time"
)

func OnDisconnection(s socketio.Conn, reason string, config *configuration.Conf) {
	fmt.Println("Disconnect from: ", s.RemoteAddr())
	fmt.Println("closed", reason)
	go func() {
		remoteAddrArr := strings.Split(s.RemoteAddr().String(), ":")
		ip, port := remoteAddrArr[0], remoteAddrArr[1]
		warmupElementData := warmup.WarmupElement{
			IP:   ip,
			Port: port,
			Status: false,
		}
		var warmupData []warmup.WarmupElement
		warmupData = append(warmupData, warmupElementData)
		warmupDataQueue := warmup.WarmupMessage{
			EventTime: time.Now().Unix(),
			Data:      warmupData,
			WupType: "event",
		}
		var warmupMsgString string
		warmupMsgString, err := warmupDataQueue.GetJsonString()
		if err != nil {
			log.Println(err)
			return
		}
		var mq messagequeue.MQ
		mq.PushMsgByTopic(config, warmupMsgString, config.MQ.WarmUpTopic)
	}()

}
