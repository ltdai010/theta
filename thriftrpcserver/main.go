package main

import (
	"fmt"
	"github.com/apache/thrift/lib/go/thrift"
	"github.com/spf13/viper"
	"log"
	"theta/cmd/thetacli/cmd/utils"
	"theta/thrift/gen-go/rpc/theta"
)

func main()  {
	protocolFactory := thrift.NewTBinaryProtocolFactory(true, true)
	var transportFactory thrift.TTransportFactory
	transportFactory = thrift.NewTBufferedTransportFactory(8192)
	transportFactory = thrift.NewTFramedTransportFactory(transportFactory)
	socket, err := thrift.NewTServerSocket("0.0.0.0:18888")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%T\n", socket)
	handler := &RpcHandler{serverAddress: viper.GetString(utils.CfgRemoteRPCEndpoint)}
	processor := theta.NewThetaServiceProcessor(handler)
	server := thrift.NewTSimpleServer4(processor, socket, transportFactory, protocolFactory)

	fmt.Println("Starting the simple server... on ", "0.0.0.0:18888")
	err = server.Serve()
	if err != nil {
		log.Fatal(err)
	}
}
