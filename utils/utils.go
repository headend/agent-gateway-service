package utils

import (
	"github.com/headend/share-module/configuration"
	static_config "github.com/headend/share-module/configuration/static-config"
	"strings"
)

func GetIpAndPortFromRemoteAddr(connAddr string) (string, string) {
	s := strings.Split(connAddr, ":")
	ip, port := s[0], s[1]
	return ip, port
}

func GetGWConnectionInfo(conf configuration.Conf) (string, uint16) {
	var gwHost string
	if conf.AgentGateway.Gateway != "" {
		gwHost = conf.AgentGateway.Gateway
	} else {
		if conf.AgentGateway.Host != "" {
			gwHost = conf.AgentGateway.Host
		} else {
			gwHost = static_config.GatewayHost
		}
	}

	var gwPort uint16
	if conf.AgentGateway.Port != 0 {
		gwPort = conf.AgentGateway.Port
	} else {
		gwPort = static_config.GatewayPort
	}
	return gwHost, gwPort
}
