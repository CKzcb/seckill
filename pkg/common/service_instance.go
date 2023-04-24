/**
 * @Author: 蛋白质先生
 * @Description:
 * @File: service_instance
 * @Version: 1.0.0
 * @Date: 2023/4/22 23:13
 */

package common

type ServiceInstance struct {
	Host      string
	Port      int
	Weight    int
	CurWeight int
	GrpcPort  int
}
