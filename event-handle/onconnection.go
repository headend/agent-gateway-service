package event_handle

import (
	"fmt"
	socketio "github.com/googollee/go-socket.io"
	messagequeue "github.com/headend/share-module/MQ"
	"github.com/headend/share-module/configuration"
	"github.com/headend/share-module/configuration/socket-event"
	"github.com/headend/share-module/model/register"
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
	nraData.IP, nraData.Port = remoteAddrArr[0], remoteAddrArr[1]
	nraData.EventTime = time.Now().Unix()
	nraMessageString, err := nraData.GetJsonString()
	if err != nil {
		log.Println(err)
		return err
	} else {
		var queueServer messagequeue.MQ
		queueServer.PushMsgByTopic(conf, nraMessageString, conf.MQ.NraTopic)
	}
	return nil
}
