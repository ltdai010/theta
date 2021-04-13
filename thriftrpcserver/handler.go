package main

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/pborman/uuid"
	"io/ioutil"
	"log"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"theta/common"
	"theta/crypto"
	"theta/ledger/types"
	"theta/thrift/gen-go/rpc/theta"
	"theta/wallet/softwallet/keystore"
)

type RpcHandler struct {
	serverAddress string
}

func (r *RpcHandler) GetAccount(ctx context.Context, address string) (*theta.Account, error) {
	body, err := r.SendRpc("{\n" +
		"\"jsonrpc\":\"2.0\",\n" +
		"\"method\":\"theta.GetAccount\",\n" +
		"\"params\":[{\"address\":\"" + address +"\"}],\n" +
		"\"id\":1\n}")
	fmt.Println(string(body))
	resultData := map[string]interface{}{}
	err = json.Unmarshal(body, &resultData)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	result := theta.AccountResult_{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	if result.Error != nil && result.Error.Code != 0 {
		return nil, errors.New(result.Error.Message)
	}

	log.Println(result)

	return result.Result_, nil
}

func (r *RpcHandler) GetTokenBalance(context context.Context, address string, contractAddress string) {
	prefix := "70a08231"
	data := prefix + "000000000000000000000000" + address

	from := types.TxInput{
		Address: common.HexToAddress(""),
		Coins: types.Coins{
			ThetaWei: new(big.Int).SetUint64(0),
			TFuelWei: big.NewInt(0),
		},
		Sequence: 0,
	}

	to := types.TxOutput{
		Address: common.HexToAddress(contractAddress),
	}

	dataHex, err := hex.DecodeString(data)
	if err != nil {
		return
	}

	smartContractTx := &types.SmartContractTx{
		From:     from,
		To:       to,
		GasLimit: 0,
		GasPrice: big.NewInt(0),
		Data:     dataHex,
	}

	smartContractTx.SignBytes("privatenet")

}

func (r *RpcHandler) SendTx(context context.Context, send *theta.Send) (*theta.BroadcastRawTransactionAsync, error) {
	if len(send.FromAddress) == 0 || len(send.To) == 0 {
		return nil, fmt.Errorf("The from and to address cannot be empty")
	}
	if send.FromAddress == send.To {
		return nil, fmt.Errorf("The from and to address cannot be identical")
	}

	from := common.HexToAddress(send.FromAddress)
	to := common.HexToAddress(send.To)

	thetawei, ok := new(big.Int).SetString(send.Thetawei, 10)
	if !ok {
		return nil, fmt.Errorf("Failed to parse thetawei: %v", send.Thetawei)
	}
	tfuelwei, ok := new(big.Int).SetString(send.Tfuelwei, 10)
	if !ok {
		return nil, fmt.Errorf("Failed to parse tfuelwei: %v", send.Tfuelwei)
	}
	fee, ok := new(big.Int).SetString(send.Fee, 10)
	if !ok {
		return nil, fmt.Errorf("Failed to parse fee: %v", send.Fee)
	}

	acc, err := r.GetAccount(context, send.FromAddress)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	seq, err := strconv.ParseInt(acc.Sequence, 10, 64)

	inputs := []types.TxInput{{
		Address: from,
		Coins: types.Coins{
			TFuelWei: new(big.Int).Add(tfuelwei, fee),
			ThetaWei: thetawei,
		},
		Sequence: uint64(seq + 1),
	}}
	outputs := []types.TxOutput{{
		Address: to,
		Coins: types.Coins{
			TFuelWei: tfuelwei,
			ThetaWei: thetawei,
		},
	}}
	sendTx := &types.SendTx{
		Fee: types.Coins{
			ThetaWei: new(big.Int).SetUint64(0),
			TFuelWei: fee,
		},
		Inputs:  inputs,
		Outputs: outputs,
	}

	decoded, err := hex.DecodeString(send.PrivateKey)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	privateKey, err := crypto.PrivateKeyFromBytes(decoded)
	if err != nil {
		log.Println("thriftrpcserver/handler.go:109", err)
		log.Println(len(decoded))
		return nil, err
	}

	signBytes := sendTx.SignBytes(send.ChainID)
	key := &keystore.Key{
		Id:         uuid.NewRandom(),
		Address:    from,
		PrivateKey: privateKey,
	}

	sig, err := key.Sign(signBytes)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	sendTx.SetSignature(from, sig)
	raw, err := types.TxToBytes(sendTx)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	signedTx := hex.EncodeToString(raw)
	msg := "{" +
		"\"jsonrpc\":\"2.0\"," +
		"\"method\":\"theta.BroadcastRawTransactionAsync\"," +
		"\"params\":[" +
		"{\"tx_bytes\":\""+ signedTx + "\"}]" +
		",\"id\":1}"
	body, err := r.SendRpc(msg)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	result := theta.BroadcastRawTransactionAsyncResult_{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	if result.Error != nil && result.Error.Code != 0 {
		return nil, errors.New(result.Error.Message)
	}

	log.Println(result)
	log.Println(string(body))

	return result.Result_, nil
}

func (r *RpcHandler) SendRpc(msg string) ([]byte, error) {
	urlAddress := r.serverAddress
	method := "POST"

	payload := strings.NewReader(msg)

	client := &http.Client {}
	req, err := http.NewRequest(method, urlAddress, payload)

	if err != nil {
		log.Println(err)
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return body, nil
}
