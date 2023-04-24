/**
 * @Author: 蛋白质先生
 * @Description:
 * @File: discovery_client
 * @Version: 1.0.0
 * @Date: 2023/4/23 07:04
 */

package discover

import (
	"github.com/CKzcb/seckill/pkg/common"
	"github.com/go-kit/kit/sd/consul"
	"github.com/hashicorp/consul/api"
	"log"
	"sync"
)

type DiscoveryClientInstance struct {
	Host   string
	Port   int
	config *api.Config
	client consul.Client
	mutex  sync.Mutex

	instancesMap sync.Map
}

type DiscoveryClient interface {
	Register(instanceId, svcHost, healthCheckUrl, svcPort string, svcName string, weight int, meta map[string]string,
		tags []string, logger *log.Logger) bool
	DeRegister(instanceId string, logger *log.Logger) bool
	DiscoverServices(serviceName string, logger *log.Logger) []*common.ServiceInstance
}
