package main

import (
	"fmt"
	"github.com/apache/thrift/lib/go/thrift"
	"log"
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
	//http://guardian-testnet-rpc.thetatoken.org:16888/rpc
	handler := &RpcHandler{serverAddress: "http://localhost:16888/rpc",
		chainID: "testnet_sapphire"}
	processor := theta.NewThetaServiceProcessor(handler)
	server := thrift.NewTSimpleServer4(processor, socket, transportFactory, protocolFactory)

	fmt.Println("Starting the simple server... on ", "0.0.0.0:18888")
	err = server.Serve()
	if err != nil {
		log.Fatal(err)
	}
}
