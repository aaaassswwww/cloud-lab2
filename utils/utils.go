package utils

import (
	"context"
	"net"
	"strings"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/pkg/discovery"
	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/pkg/utils"
)

var (
	CartHostPorts     = []string{"cart:8883"}
	CheckoutHostPorts = []string{"checkout:8884"}
	EmailHostPorts    = []string{"email:8888"}
	OrderHostPorts    = []string{"order:8885"}
	PaymentHostPorts  = []string{"payment:8886"}
	ProductHostPorts  = []string{"product:8881"}
	UserHostPorts     = []string{"user:8882"}
)

func MyWithHostPorts(hostports ...string) client.Option {
	return client.Option{F: func(o *client.Options, di *utils.Slice) {
		o.Targets = strings.Join(hostports, ",")
		o.Resolver = &discovery.SynthesizedResolver{
			ResolveFunc: func(ctx context.Context, key string) (discovery.Result, error) {
				var ins []discovery.Instance
				for _, hp := range hostports {
					if _, err := net.ResolveTCPAddr("tcp", hp); err == nil {
						ins = append(ins, discovery.NewInstance("tcp", hp, discovery.DefaultWeight, nil))
						continue
					}
					if _, err := net.ResolveUnixAddr("unix", hp); err == nil {
						ins = append(ins, discovery.NewInstance("unix", hp, discovery.DefaultWeight, nil))
						continue
					}
				}
				return discovery.Result{
					Cacheable: true,
					CacheKey:  "fixed",
					Instances: ins,
				}, nil
			},
			NameFunc: func() string { return o.Targets },
			TargetFunc: func(ctx context.Context, target rpcinfo.EndpointInfo) string {
				return o.Targets
			},
		}
	}}
}