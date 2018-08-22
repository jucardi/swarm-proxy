package proxy

import (
	"github.com/jucardi/swarm-proxy/docker"
	"github.com/jucardi/swarm-proxy/model"
	"github.com/jucardi/go-streams/streams"
	"github.com/docker/docker/api/types/swarm"
	"strconv"
	"github.com/jucardi/go-logger-lib/log"
	"strings"
	"github.com/jucardi/go-beans/beans"
	"github.com/jucardi/infuse/templates"
	"github.com/jucardi/infuse/util/ioutils"
	"os"
	"github.com/jucardi/swarm-proxy/config"
	"fmt"
)

const (
	PublishKey     = "com.jucardi.swarm.proxy.publish"
	ProxyPortKey   = "com.jucardi.swarm.proxy.port_map"
	ProxyUriKey    = "com.jucardi.swarm.proxy.location"
	RedirectUriKey = "com.jucardi.swarm.proxy.redirect_map"
	RewriteKey     = "com.jucardi.swarm.proxy.rewrite"
	ServerNameKey  = "com.jucardi.swarm.proxy.server_name"

	DefaultBeanName = "docker-proxy-service"
)

var (
	// To validate the interface implementation at compile time instead of runtime.
	_ IProxyService = (*service)(nil)

	instance *service
)

type service struct {
}

func Service() IProxyService {
	return beans.Resolve((*IProxyService)(nil), DefaultBeanName).(IProxyService)
}

// Registering the bean implementation.
func init() {
	beans.RegisterFunc((*IProxyService)(nil), DefaultBeanName, func() interface{} {
		if instance != nil {
			return instance
		}

		instance = &service{}
		return instance
	})
}

func (s *service) ParseTemplate(info ...*model.ProxyConfig) (string, error) {
	var cInfo *model.ProxyConfig

	if len(info) > 0 {
		cInfo = info[0]
	} else {
		if ci, err := s.GetProxyConfig(); err == nil {
			cInfo = ci
		} else {
			return "", err
		}
	}

	template, err := templates.Factory().Create(templates.TypeGo)
	if err != nil {
		return "", err
	}
	if err := template.LoadFileTemplate("./fixtures/nginx.tmpl"); err != nil {
		return "", err
	}
	writer := ioutils.NewStringWriter()
	if err := template.Parse(os.Stdout, cInfo); err != nil {
		return "", err
	}
	return writer.ToString(), nil
}

func (s *service) GetProxyConfig() (*model.ProxyConfig, error) {
	services, err := docker.Client().GetServices()
	if err != nil {
		return nil, err
	}
	//nodes, err := docker.Client().GetNodes()
	//if err != nil {
	//	return nil, err
	//}

	ret := model.NewProxyConfig()
	streams.From(services).

		Filter(func(i interface{}) bool { // Filter the services that only have the minimum necessary params. TODO: Automatically get this filtered from the API using the args
			service := i.(swarm.Service)
			_, ok1 := service.Spec.Labels[PublishKey]
			_, ok2 := service.Spec.Labels[ProxyPortKey]
			_, ok3 := service.Spec.Labels[ProxyUriKey]
			if ok1 && (!ok2 || !ok3) {
				log.Warnf("Unable to process service '%s', incomplete data, requires `port_map` and `location`", service.Spec.Name)
			}
			return ok1 && ok2 && ok3
		}).

		Filter(func(i interface{}) bool { // Client only services with the publish label true. TODO: Automatically get this filtered from the API using the args
			service := i.(swarm.Service)
			val := service.Spec.Labels[PublishKey]
			ret, err := strconv.ParseBool(val)
			if err != nil {
				log.Warnf("[%s] Error parsing value '%s', '%v'", service.Spec.Name, val, err)
			}
			log.Debugf("Publish %s: %v", service.Spec.Name, ret)
			return ret
		}).

		ForEach(func(i interface{}) {
			service := i.(swarm.Service)
			mapServiceModeAll(service, ret)
			mapServiceModeLabel(service, ret)
		})

	return ret, nil
}

func mapServiceModeAll(service swarm.Service, cfg *model.ProxyConfig) {
	if config.Get().Mode()&config.ModeAll != config.ModeAll {
		return
	}

	ports := service.Spec.EndpointSpec.Ports
	name := service.Spec.Name

	if len(ports) == 0 {
		log.Warnf("no port mapping for service %s, unable to determine port to proxy, skipping", name)
		return
	}
	containerPort := int(ports[0].TargetPort)
	proxyPort := config.Get().DefaultProxyPort
	if len(ports) > 1 {
		log.Warnf("multiple ports mapping for service %s, using the first one in the list %d:%d", name, ports[0].PublishedPort, ports[0].TargetPort)
	}

	var server *model.ServerInfo

	upstream := cfg.Upstreams.Set(name, containerPort)

	if cfg.Servers.Contains("", proxyPort) {
		server = cfg.Servers.Get("", proxyPort)
	} else {
		server = &model.ServerInfo{
			Name:       "Gateway",
			ListenPort: proxyPort,
		}
		cfg.Servers.Set("", proxyPort, server)
	}

	rewrite := fmt.Sprintf("/%s/(.*) /$1 break", name)
	server.AddLocation(model.LocationInfo{
		Rewrite:   &rewrite,
		Location:  fmt.Sprintf("/%s/", name),
		ProxyPass: upstream.Name,
	})
}

func mapServiceModeLabel(service swarm.Service, cfg *model.ProxyConfig) {
	if config.Get().Mode()&config.ModeLabels != config.ModeLabels {
		return
	}

	name := service.Spec.Name
	portSplit := strings.Split(service.Spec.Labels[ProxyPortKey], ":")
	uriLocation := service.Spec.Labels[ProxyUriKey]

	if len(portSplit) != 2 {
		log.Warnf("[%s] Unexpected argument count for `port_map`. Port mapping need to be in the format of '[service_port]:[publish_port]'", service.Spec.Name)
		return
	}

	publishPort, err := strconv.Atoi(portSplit[0])
	if err != nil {
		log.Warnf("[%s] Error parsing publish port: '%s'", service.Spec.Name, portSplit[0])
		return
	}
	servicePort, err := strconv.Atoi(portSplit[1])
	if err != nil {
		log.Warnf("[%s] Error parsing service port: '%s'", service.Spec.Name, portSplit[1])
		return
	}

	var server *model.ServerInfo

	upstream := cfg.Upstreams.Set(name, servicePort)
	serverName := ""

	if val, ok := service.Spec.Labels[ServerNameKey]; ok {
		serverName = val
	}

	// TODO: Allow using proxy port by default
	if cfg.Servers.Contains(serverName, publishPort) {
		server = cfg.Servers.Get(serverName, publishPort)
		server.Name = fmt.Sprintf("%s, $s", server.Name, name)
	} else {
		server = &model.ServerInfo{
			Name:       name,
			ListenPort: publishPort,
		}
		cfg.Servers.Set(serverName, publishPort, server)
	}

	if serverName != "" {
		server.ServerName = &serverName
	}

	location := model.LocationInfo{
		Location:  uriLocation,
		ProxyPass: upstream.Name,
	}

	if val, ok := service.Spec.Labels[RewriteKey]; ok {
		location.Rewrite = &val
	}

	server.AddLocation(location)

	redirect, ok := service.Spec.Labels[RedirectUriKey]
	if !ok {
		return
	}

	redirectSplit := strings.Split(redirect, ":")
	if len(portSplit) != 2 {
		log.Warnf("[%s] Unexpected argument count for `redirect_map`. Redirect mapping need to be in the format of '[source_uri]:[target_uri]'", service.Spec.Name)
		return
	}

	server.AddRedirect(model.RedirectInfo{
		Source: redirectSplit[0],
		Target: redirectSplit[1],
	})
}
