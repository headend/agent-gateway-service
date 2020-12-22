package utils

import (
	"context"
	agentpb "github.com/headend/iptv-agent-service/proto"
	"github.com/headend/share-module/mygrpc/connection"
	"time"
)

func GetAgentByID(agentServerHost string, agentServerPort uint16, err error, agentid int64) *agentpb.AgentResponse {
	var rpcConnection connection.RpcClient
	rpcConnection.InitializeClient(agentServerHost, agentServerPort)
	defer rpcConnection.Client.Close()
	agentClient := agentpb.NewAgentCTLServiceClient(rpcConnection.Client)
	// Check Agent exists
	c, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	res, err := (agentClient).Get(c, &agentpb.AgentFilter{Id: agentid})
	return res
}
