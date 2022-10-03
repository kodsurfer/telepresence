package managerutil

import (
	"context"
	"encoding/json"
	"net"
	"strings"
	"time"

	"github.com/sethvargo/go-envconfig"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/telepresenceio/telepresence/v2/pkg/agentconfig"
	"github.com/telepresenceio/telepresence/v2/pkg/agentmap"
	"github.com/telepresenceio/telepresence/v2/pkg/k8sapi"
	"github.com/telepresenceio/telepresence/v2/pkg/version"
)

type Env struct {
	LogLevel       string `env:"LOG_LEVEL,default=info"`
	User           string `env:"USER,default="`
	ServerHost     string `env:"SERVER_HOST,default="`
	ServerPort     int32  `env:"SERVER_PORT,default=8081"`
	PrometheusPort string `env:"PROMETHEUS_PORT,default=0"`
	SystemAHost    string `env:"SYSTEMA_HOST,default="`
	SystemAPort    int32  `env:"SYSTEMA_PORT,default=0"`

	ManagerNamespace    string                     `env:"MANAGER_NAMESPACE,default="`
	ManagedNamespaces   string                     `env:"MANAGED_NAMESPACES,default="`
	AgentRegistry       string                     `env:"AGENT_REGISTRY,default=docker.io/datawire"`
	AgentImage          string                     `env:"AGENT_IMAGE,default="`
	AgentPort           int32                      `env:"AGENT_PORT,default=9900"`
	AgentResources      string                     `env:"AGENT_RESOURCES,default="`
	AgentInitResources  string                     `env:"AGENT_INIT_RESOURCES,default="`
	APIPort             int32                      `env:"AGENT_REST_API_PORT,default="`
	TracingGrpcPort     int32                      `env:"TRACING_GRPC_PORT,default="`
	MaxReceiveSize      resource.Quantity          `env:"GRPC_MAX_RECEIVE_SIZE,default=4Mi"`
	AppProtocolStrategy k8sapi.AppProtocolStrategy `env:"AGENT_APP_PROTO_STRATEGY,default="`
	AgentInjectPolicy   agentconfig.InjectPolicy   `env:"AGENT_INJECT_POLICY,default="`

	PodCIDRStrategy string `env:"POD_CIDR_STRATEGY,default=auto"`
	PodCIDRs        string `env:"POD_CIDRS,default="`
	PodIP           string `env:"POD_IP,default="`

	DNSServiceName      string `env:"DNS_SERVICE_NAME,default="`
	DNSServiceNamespace string `env:"DNS_SERVICE_NAMESPACE,default="`
	DNSServiceIP        string `env:"DNS_SERVICE_IP,default="`

	ClientRoutingAlsoProxySubnets  []string      `env:"CLIENT_ROUTING_ALSO_PROXY_SUBNETS,default="`
	ClientRoutingNeverProxySubnets []string      `env:"CLIENT_ROUTING_NEVER_PROXY_SUBNETS,default="`
	ClientDnsExcludeSuffixes       []string      `env:"CLIENT_DNS_EXCLUDE_SUFFIXES,default=.com,.io,.net,.org,.ru"`
	ClientDnsIncludeSuffixes       []string      `env:"CLIENT_DNS_INCLUDE_SUFFIXES,default="`
	ClientConnectionTTL            time.Duration `env:"CLIENT_CONNECTION_TTL,default=24h"`
}

type envKey struct{}

func (e *Env) GeneratorConfig(qualifiedAgentImage string) (*agentmap.GeneratorConfig, error) {
	gc := &agentmap.GeneratorConfig{
		AgentPort:           uint16(e.AgentPort),
		APIPort:             uint16(e.APIPort),
		TracingPort:         uint16(e.TracingGrpcPort),
		ManagerPort:         uint16(e.ServerPort),
		QualifiedAgentImage: qualifiedAgentImage,
		ManagerNamespace:    e.ManagerNamespace,
		LogLevel:            e.LogLevel,
	}
	parseResources := func(js string) (*core.ResourceRequirements, error) {
		if js == "" {
			return nil, nil
		}
		var rr *core.ResourceRequirements
		if err := json.Unmarshal([]byte(js), &rr); err != nil {
			return nil, err
		}
		return rr, nil
	}
	var err error
	if gc.InitResources, err = parseResources(e.AgentInitResources); err != nil {
		return nil, err
	}
	if gc.Resources, err = parseResources(e.AgentResources); err != nil {
		return nil, err
	}
	return gc, nil
}

func (e *Env) QualifiedAgentImage() string {
	img := e.AgentImage
	if img == "" {
		img = "tel2:" + strings.TrimPrefix(version.Version, "v")
	}
	return e.AgentRegistry + "/" + img
}

func (e *Env) GetManagedNamespaces() []string {
	if mns := e.ManagedNamespaces; mns != "" {
		return strings.Split(mns, " ")
	}
	return nil
}

func (e *Env) GetAlsoProxySubnets() ([]*net.IPNet, error) {
	return parseRawSubnets(e.ClientRoutingAlsoProxySubnets)
}

func (e *Env) GetNeverProxySubnets() ([]*net.IPNet, error) {
	return parseRawSubnets(e.ClientRoutingNeverProxySubnets)
}

func parseRawSubnets(ipNetStrs []string) ([]*net.IPNet, error) {
	if len(ipNetStrs) == 0 { // env var not set
		return nil, nil
	}

	ipNets := make([]*net.IPNet, len(ipNetStrs))
	for i, s := range ipNetStrs {
		_, ipNet, err := net.ParseCIDR(s)
		if err != nil {
			return nil, err
		}
		ipNets[i] = ipNet
	}
	return ipNets, nil
}

func LoadEnv(ctx context.Context) (context.Context, error) {
	var env Env
	if err := envconfig.Process(ctx, &env); err != nil {
		return ctx, err
	}
	return WithEnv(ctx, &env), nil
}

func WithEnv(ctx context.Context, env *Env) context.Context {
	return context.WithValue(ctx, envKey{}, env)
}

func GetEnv(ctx context.Context) *Env {
	env, ok := ctx.Value(envKey{}).(*Env)
	if !ok {
		return nil
	}
	return env
}
