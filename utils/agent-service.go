package utils

import (
	"context"
	agentpb "github.com/headend/iptv-agent-service/proto"
	"github.com/headend/share-module/configuration"
	agentModel "github.com/headend/share-module/model/agentd"
	"github.com/headend/share-module/mygrpc/connection"
	"time"
)

func GetAgentByID(agentServerHost string, agentServerPort uint16, err error, agentid int64) *agentpb.AgentResponse {
	var rpcConnection connection.RpcClient
	rpcConnection.InitializeClient(agentServerHost, agentServerPort)
	defer rpcConnection.Client.Close()
	agentClient := agentpb.NewAgentServiceClient(rpcConnection.Client)
	// Check Agent exists
	c, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	res, err := (agentClient).Get(c, &agentpb.AgentFilter{Id: agentid})
	return res
}

func GetAgentByIP(agentServerHost string, agentServerPort uint16, agentdIP string) *agentpb.Agent {
	var rpcConnection connection.RpcClient
	rpcConnection.InitializeClient(agentServerHost, agentServerPort)
	defer rpcConnection.Client.Close()
	agentClient := agentpb.NewAgentServiceClient(rpcConnection.Client)
	// Check Agent exists
	c, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	res, err := (agentClient).Get(c, &agentpb.AgentFilter{IpControl: agentdIP})
	if err != nil {
		println(err)
		return nil
	}
	if len(res.Agents) == 1 {
		return res.Agents[0]
	} else {
		return nil
	}
}

func GetProfileMonitorByAgent(conf configuration.Conf, ip string, monitorType int) (*agentpb.ProfileMonitorResponse, error) {
	var agentRpcconn connection.RpcClient
	agentRpcconn.InitializeClient(conf.RPC.Agent.Gateway, conf.RPC.Agent.Port)
	defer agentRpcconn.Client.Close()
	agentCli := agentpb.NewAgentServiceClient(agentRpcconn.Client)
	c, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	res, err := (agentCli).GetProfileMonitor(c, &agentpb.ProfileMonitorRequest{IpControl: ip, MonitorType: int64(monitorType)})
	return res, err
}

func OnMonitorChangeUpdateStatus(conf configuration.Conf, onProfileChangeStatus agentModel.ProfileChangeStatus) error {
	var rpcConnection connection.RpcClient
	rpcConnection.InitializeClient(conf.RPC.Agent.Gateway, conf.RPC.Agent.Port)
	defer rpcConnection.Client.Close()
	agentClient := agentpb.NewAgentServiceClient(rpcConnection.Client)
	// Check Agent exists
	c, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_, err2 := (agentClient).MonitorUpdateStatus(c, &agentpb.MonitorUpdateStatusRequest{
		MonitorId:       onProfileChangeStatus.MonitorID,
		MonitorType:     int64(onProfileChangeStatus.MonitorType),
		NewStatus:       onProfileChangeStatus.NewStatus,
		NewSignalStatus: onProfileChangeStatus.NewSignalStatus,
		NewVideoStatus:  onProfileChangeStatus.NewVideoStatus,
		NewAudioStatus:  onProfileChangeStatus.NewAudioStatus,
	})
	return err2
}