package broadcast_to_agentd

import (
	"encoding/json"
	"fmt"
	socket_event "github.com/headend/share-module/configuration/socket-event"
	file_and_directory "github.com/headend/share-module/file-and-directory"
	share_model "github.com/headend/share-module/model"
	"github.com/googollee/go-socket.io"
	"log"
)

func BroadcastFileToRom(filepath string, server *socketio.Server) {
	/*
		Send file info to rom
	*/
	// get md5 from file
	md5str, err := file_and_directory.GetMd5FromFile("asset/" + filepath)
	if err != nil {
		panic(err)
	} else {
		fileSize, err := file_and_directory.GetFileSizeInByte("asset/" + filepath)
		if err != nil {
			panic(err)
		} else {
			fileInfo2Send := share_model.WorkerUpdateSignal{
				FileName:       filepath,
				FilePath:       filepath,
				FileSizeInByte: fileSize,
				Md5:            md5str,
			}
			b, err := json.Marshal(fileInfo2Send)
			if err != nil {
				fmt.Printf("Error: %s", err)
			}
			server.BroadcastToRoom("/", "monitor", socket_event.NhanFile, string(b))
			log.Printf("File info to send:%v\n", string(b))
		}
	}
}
