package main

import (
	"context"
	"encoding/hex"
	"encoding/json"
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
	err = json.Unmarshal(body, &theta.AccountResult_{})
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return result.Result_, nil
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
		Sequence: uint64(seq),
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

	privateKey, err := crypto.PrivateKeyFromBytes([]byte(send.PrivateKey))
	if err != nil {
		log.Println(err)
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
