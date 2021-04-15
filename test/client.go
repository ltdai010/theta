package main

import (
	"context"
	"encoding/json"
	"log"
	"theta/test/transport"
	"theta/thrift/gen-go/rpc/theta"
)

func main() {
	TestGetBlockByHeight()
}

const (
	host = "0.0.0.0"
	port = "18888"
)

func TestSend() {
	r, err := transport.GetThetaClient(host, port).Client.(*theta.ThetaServiceClient).
		SendTx(context.Background(), &theta.Send{
		PrivateKey: "93A90EA508331DFDF27FB79757D4250B4E84954927BA0073CD67454AC432C737",
		To:         "0d2fd67d573c8ecb4161510fc00754d64b401f86",
		Thetawei:   "0",
		Tfuelwei:   "10",
		Fee:        "1000000000000",
	})
	if err != nil {
		log.Fatal(err)
	}
	b, _ := json.MarshalIndent(r, "", "    ")
	log.Println(string(b))
}

func TestGetAccount() {
	r, err := transport.GetThetaClient(host, port).Client.(*theta.ThetaServiceClient).
		GetAccount(context.Background(), "2E833968E5bB786Ae419c4d13189fB081Cc43bab")
	if err != nil {
		log.Fatal(err)
	}
	b, _ := json.MarshalIndent(r, "", "    ")
	log.Println(string(b))
}

func TestGetTokenBalance() {
	r, err := transport.GetThetaClient(host, port).Client.(*theta.ThetaServiceClient).
		GetTokenBalance(context.Background(), "2e833968e5bb786ae419c4d13189fb081cc43bab",
			"413682f3ec6504695ef2d70cda502c0489ce86af",
			"93A90EA508331DFDF27FB79757D4250B4E84954927BA0073CD67454AC432C737")
	if err != nil {
		log.Fatal(err)
	}
	b, err := json.MarshalIndent(r, "", "\t")
	log.Println(string(b))
}

func TestGetBlockByHash() {
	r, err := transport.GetThetaClient(host, port).Client.(*theta.ThetaServiceClient).
		GetBlock(context.Background(), "")
	if err != nil {
		log.Fatal(err)
	}
	b, err := json.MarshalIndent(r, "", "\t")
	log.Println(string(b))
}

func TestGetBlockByHeight() {
	r, err := transport.GetThetaClient(host, port).Client.(*theta.ThetaServiceClient).
		GetBlockByHeight(context.Background(), 3)
	if err != nil {
		log.Fatal(err)
	}
	b, err := json.MarshalIndent(r, "", "\t")
	log.Println(string(b))
}