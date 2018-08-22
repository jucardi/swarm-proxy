package model

import (
	"fmt"
)

type ProxyConfig struct {
	Upstreams UpstreamsMap
	Servers   ServersMap
}

type ServerInfo struct {
	Name       string
	ListenPort int
	ServerName *string
	Locations  []LocationInfo
	Redirects  []RedirectInfo
}

type LocationInfo struct {
	Rewrite   *string
	Location  string
	ProxyPass string
}

type RedirectInfo struct {
	Source string
	Target string
}

type UpstreamInfo struct {
	Name     string
	Hostname string
	Port     int
}

type ServersMap map[string]*ServerInfo

func (m ServersMap) Contains(serverName string, port int) bool {
	_, ok := m[m.key(serverName, port)]
	return ok
}

func (m ServersMap) Get(serverName string, port int) *ServerInfo {
	val := m[m.key(serverName, port)]
	return val
}

func (m ServersMap) Set(serverName string, port int, val *ServerInfo) {
	m[m.key(serverName, port)] = val
}

func (m ServersMap) key(serverName string, port int) string {
	return fmt.Sprintf("%s:%v", serverName, port)
}

type UpstreamsMap map[string]*UpstreamInfo

func (m UpstreamsMap) Contains(serviceName string, port int) bool {
	_, ok := m[m.key(serviceName, port)]
	return ok
}

func (m UpstreamsMap) Get(serviceName string, port int) *UpstreamInfo {
	val := m[m.key(serviceName, port)]
	return val
}

func (m UpstreamsMap) Set(serviceName string, port int) *UpstreamInfo {
	name := m.key(serviceName, port)
	val := &UpstreamInfo{
		Name:     name,
		Hostname: serviceName,
		Port:     port,
	}
	m[name] = val
	return val
}

func (m UpstreamsMap) key(serviceName string, port int) string {
	return fmt.Sprintf("%s_%v", serviceName, port)
}

func (s *ServerInfo) AddLocation(loc LocationInfo) {
	s.Locations = append(s.Locations, loc)
}

func (s *ServerInfo) AddRedirect(redirect RedirectInfo) {
	s.Redirects = append(s.Redirects, redirect)
}

func NewProxyConfig() *ProxyConfig {
	return &ProxyConfig{
		Upstreams: map[string]*UpstreamInfo{},
		Servers:   map[string]*ServerInfo{},
	}
}
