package config

const (
	// ModeLabel indicates that only the services with proxy labels will be published
	ModeLabels = 0x01

	// ModeAll indicates that all services will be published as [service name]:[service port]/* to [proxy ingress]:[proxy ingress port]/[service name]/*
	ModeAll = 0x02

	// ModeMixed indicates that all services will be published like `ModeAll` but also services with proxy label config will also get published using the provided configuration
	ModeMixed = ModeLabels | ModeAll
)

var modeMap = map[string]PublishMode{
	"labels": ModeLabels,
	"all":    ModeAll,
	"mixed":  ModeMixed,
}
// PublishMode indicates the publish mode
type PublishMode int

// Config contains the configuration the proxy service will use to run
type Config struct {
	// Mode indicates the publish mode
	PublishMode string `json:"publish_mode" yaml:"publish_mode"`
	// DefaultProxyPort indicates the port to use in `Mode=all` or when a port is not specified in the label configuration
	DefaultProxyPort int `json:"default_proxy_port" yaml:"default_proxy_port"`
}

var instance *Config

func Get() *Config {
	if instance == nil {
		instance = &Config{
			PublishMode:      "mixed",
			DefaultProxyPort: 8765,
		}
	}
	return instance
}

func (c *Config) Mode() PublishMode {
	return modeMap[c.PublishMode] // Panic is intentional if the key is not found, it means an invalid mode was configured
}
