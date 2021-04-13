package main

import (
	"context"
	"log"
	"theta/test/transport"
	"theta/thrift/gen-go/rpc/theta"
)

func main() {
	r, err := transport.GetThetaClient("0.0.0.0", "18888").Client.(*theta.ThetaServiceClient).
		SendTx(context.Background(), &theta.Send{
		ChainID:     "privatenet",
		FromAddress: "0x2E833968E5bB786Ae419c4d13189fB081Cc43bab",
		To:          "0x06CC5fcf74643381531773B938B9c4a2973F6eA6",
		Thetawei:    "20",
		Tfuelwei:    "20",
		Fee:         "1000000000000",
		PrivateKey:  "93A90EA508331DFDF27FB79757D4250B4E84954927BA0073CD67454AC432C737",
	})
	if err != nil {
		log.Fatal(err)
	}
	log.Println(r)
}


