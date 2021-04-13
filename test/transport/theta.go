package transport

import (
	thriftpool "github.com/OpenStars/thriftpoolv2"
	"github.com/apache/thrift/lib/go/thrift"
	"theta/thrift/gen-go/rpc/theta"
)


var (
	thetaServiceCompactMapPool = thriftpool.NewMapPool(1000, 3600, 3600,
		thriftpool.GetThriftClientCreatorFunc(func(c thrift.TClient) interface{} {
			return (theta.NewThetaServiceClient(c)) }),
		thriftpool.DefaultClose)
)

func GetThetaClient(aHost, aPort string) *thriftpool.ThriftSocketClient {
	client, _ := thetaServiceCompactMapPool.Get(aHost, aPort).Get()
	return client
}
