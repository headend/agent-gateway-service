package utils

import (
	"context"
	"fmt"
	agentpb "github.com/headend/iptv-agent-service/proto"
	loggingpb "github.com/headend/iptv-logging-service/proto"
	"github.com/headend/share-module/configuration"
	agentModel "github.com/headend/share-module/model/agentd"
	"github.com/headend/share-module/mygrpc/connection"
	"log"
	"time"
)

func DoWriNonitorLog(conf configuration.Conf, logData loggingpb.MonitorLogsRequest) {
	log.Println(logData)
	var rpcConnection connection.RpcClient
	rpcConnection.InitializeClient(conf.RPC.Logging.Gateway, conf.RPC.Logging.Port)
	defer rpcConnection.Client.Close()
	loggingClient := loggingpb.NewMonitorLogsServiceClient(rpcConnection.Client)
	// Check Agent exists
	c, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_, err3 := (loggingClient).Add(c, &logData)
	if err3 != nil {
		log.Println(err3)
	}
}

func MakeLogInDataRequest(onProfileChangeStatus agentModel.ProfileChangeStatus, monitorInfo *agentpb.ProfileMonitorElement) loggingpb.MonitorLogsRequest {
	var desc string
	switch onProfileChangeStatus.NewStatus {
	case 0:
		desc = fmt.Sprintf("Channel %s with ip %s DOWN on host %s id %s", monitorInfo.ChannelName, monitorInfo.MulticastIp, monitorInfo.IpControl, monitorInfo.AgentId)
	case 1:
		desc = fmt.Sprintf("Channel %s with ip %s UP on host %s id%s", monitorInfo.ChannelName, monitorInfo.MulticastIp, monitorInfo.IpControl, monitorInfo.AgentId)
	case 2:
		desc = fmt.Sprintf("Channel %s with ip %s NoVideo on host %s id %s", monitorInfo.ChannelName, monitorInfo.MulticastIp, monitorInfo.IpControl, monitorInfo.AgentId)
	case 3:
		desc = fmt.Sprintf("Channel %s with ip %s NoAudio on host %s id %s", monitorInfo.ChannelName, monitorInfo.MulticastIp, monitorInfo.IpControl, monitorInfo.AgentId)
	default:
		desc = fmt.Sprintf("Channel %s with ip %s Unknow on host %s id %s", monitorInfo.ChannelName, monitorInfo.MulticastIp, monitorInfo.IpControl, monitorInfo.AgentId)
	}
	logData := loggingpb.MonitorLogsRequest{
		AgentId:            monitorInfo.AgentId,
		ProfileId:          monitorInfo.ProfileId,
		MonitorId:          monitorInfo.MonitorId,
		ChannelId:          monitorInfo.ChannelId,
		ChannelName:        monitorInfo.ChannelName,
		MulticastIp:        monitorInfo.MulticastIp,
		BeforeStatus:       onProfileChangeStatus.OldStatus,
		BeforeSignalStatus: onProfileChangeStatus.OldSignalStatus,
		BeforeVideoStatus:  onProfileChangeStatus.OldVideoStatus,
		BeforeAudioStatus:  onProfileChangeStatus.OldAudioStatus,
		AfterStatus:        onProfileChangeStatus.NewStatus,
		AfterSignalStatus:  onProfileChangeStatus.NewSignalStatus,
		AfterVideoStatus:   onProfileChangeStatus.NewVideoStatus,
		AfterAudioStatus:   onProfileChangeStatus.NewAudioStatus,
		Description:        desc,
	}
	return logData
}

