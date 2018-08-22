package proxy

import "github.com/jucardi/swarm-proxy/model"

type IProxyService interface {
	GetProxyConfig() (*model.ProxyConfig, error)
	ParseTemplate(info ...*model.ProxyConfig) (string, error)
}
