package event_handle

import (
	"fmt"
	socketio "github.com/googollee/go-socket.io"
)

func OnErr(s socketio.Conn, e error) (int, error) {
	s.Close()
	return fmt.Printf("Close connection from %s meet error: %s", s.RemoteAddr().String(), e.Error())
}
