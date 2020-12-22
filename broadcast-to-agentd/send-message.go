package broadcast_to_agentd

import (
	socketio "github.com/googollee/go-socket.io"
	"log"
)

func SendMessageToRom(server *socketio.Server, rom string, event string, message string) {
	server.BroadcastToRoom("/", rom, event, message)
	log.Printf("Client connected:%v\n", server.Count())
}
