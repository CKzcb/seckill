/**
 * @Author: 蛋白质先生
 * @Description:
 * @File: kit_consul_client
 * @Version: 1.0.0
 * @Date: 2023/4/23 23:17
 */

package discover

import (
	"github.com/CKzcb/seckill/pkg/common"
	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/api/watch"
	"log"
	"strconv"
)

type KitDiscoverClientInstance struct {
	DiscoveryClientInstance
}

func (client *KitDiscoverClientInstance) Register(instanceId, svcHost, healthCheckUrl, svcPort string, svcName string,
	weight int, meta map[string]string, tags []string, logger *log.Logger) bool {
	port, _ := strconv.Atoi(svcPort)
	// 1. 构建服务实例原数据
	serviceRegistration := &api.AgentServiceRegistration{
		ID:      instanceId,
		Name:    svcName,
		Port:    port,
		Address: svcHost,
		Weights: &api.AgentWeights{Passing: weight},
		Check: &api.AgentServiceCheck{
			DeregisterCriticalServiceAfter: "30s",
			HTTP:                           "http://" + svcHost + ":" + strconv.Itoa(port) + healthCheckUrl,
			Interval:                       "15s",
		},
	}
	// 2. 发送服务到consul
	err := client.client.Register(serviceRegistration)
	if err != nil {
		if logger != nil {
			logger.Println("Register Service Error ! ", err)
		}
		return false
	}
	if logger != nil {
		logger.Println("Register Service Success ! ")
	}
	return true
}

func (client *KitDiscoverClientInstance) DeRegister(instanceId string, logger *log.Logger) bool {
	// 构建元数据
	serviceRegistration := &api.AgentServiceRegistration{ID: instanceId}
	// 发送注销请求
	err := client.client.Deregister(serviceRegistration)
	if err != nil {
		if logger != nil {
			logger.Println("Deregister Service Error ! ", err)
		}
		return false
	}
	if logger != nil {
		logger.Println("Deregister Service Success ! ")
	}
	return true
}

func (client *KitDiscoverClientInstance) DiscoverServices(serviceName string, logger *log.Logger) []*common.ServiceInstance {
	// 先查看是否缓存
	instanceList, ok := client.instancesMap.Load(serviceName)
	if ok {
		return instanceList.([]*common.ServiceInstance)
	}
	// 获取
	client.mutex.Lock()
	defer client.mutex.Unlock()
	// 检查
	instanceList, ok = client.instancesMap.Load(serviceName)
	if ok {
		return instanceList.([]*common.ServiceInstance)
	} else {
		go func() {
			params := make(map[string]interface{})
			params["type"] = "service"
			params["service"] = serviceName
			plan, _ := watch.Parse(params)
			plan.Handler = func(u uint64, i interface{}) {
				if i == nil {
					return
				}
				v, ok := i.([]*api.ServiceEntry)
				if !ok {
					return
				}
				if len(v) == 0 {
					client.instancesMap.Store(serviceName, []*common.ServiceInstance{})
				}
				var healthServices []*common.ServiceInstance
				for _, service := range v {
					if service.Checks.AggregatedStatus() == api.HealthPassing {
						healthServices = append(healthServices, newServiceInstance(service.Service))
					}
				}
				client.instancesMap.Store(serviceName, healthServices)
			}
			defer plan.Stop()
			_ = plan.Run(client.config.Address)
		}()
	}
	// 获取
	entries, _, err := client.client.Service(serviceName, "", false, nil)
	if err != nil {
		client.instancesMap.Store(serviceName, []*common.ServiceInstance{})
		if logger != nil {
			logger.Println("Discover Service Error ! ", err)
		}
		return nil
	}
	instances := make([]*common.ServiceInstance, len(entries))
	for i := 0; i < len(entries); i++ {
		instances[i] = newServiceInstance(entries[i].Service)
	}
	client.instancesMap.Store(serviceName, instances)
	return instances
}

func newServiceInstance(service *api.AgentService) *common.ServiceInstance {
	rpcPort := service.Port - 1
	if service.Meta != nil {
		if rpcPortString, ok := service.Meta["rpcPort"]; ok {
			rpcPort, _ = strconv.Atoi(rpcPortString)
		}
	}
	return &common.ServiceInstance{
		Host:     service.Address,
		Port:     service.Port,
		Weight:   service.Weights.Passing,
		GrpcPort: rpcPort,
	}
}
