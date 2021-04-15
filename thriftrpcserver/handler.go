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
	"theta/cmd/thetacli/cmd/utils"
	"theta/common"
	"theta/crypto"
	"theta/ledger/types"
	"theta/thrift/gen-go/rpc/theta"
	"theta/wallet/softwallet/keystore"
)

type RpcHandler struct {
	serverAddress string
	chainID       string
}

func (r *RpcHandler) GetBlockHeader(ctx context.Context, hash string) (*theta.BlockHeader, error) {
	body, err := r.SendRpc("{\"jsonrpc\":\"2.0\",\"method\":\"theta.GetBlock\"," +
		"\"params\":[" +
		"{\"hash\":\"" + hash + "\"}]," +
		"\"id\":1}")
	if err != nil {
		return nil, err
	}

	result := theta.BlockResult_{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	if result.Error != nil && result.Error.Code != 0 {
		return nil, errors.New(result.Error.Message)
	}

	return &theta.BlockHeader{
		ChainID:          result.Result_.ChainID,
		Epoch:            result.Result_.Epoch,
		Height:           result.Result_.Height,
		Parent:           result.Result_.Parent,
		TransactionsHash: result.Result_.TransactionsHash,
		StateHash:        result.Result_.StateHash,
		Timestamp:        result.Result_.Timestamp,
		Proposer:         result.Result_.Proposer,
		Children:         result.Result_.Children,
		Status:           result.Result_.Status,
		Hash:             result.Result_.Hash,
		Hcc:              result.Result_.Hcc,
	}, nil
}

func (r *RpcHandler) GetBlockHeaderByHeight(ctx context.Context, height int64) (*theta.BlockHeader, error) {
	body, err := r.SendRpc("{\"jsonrpc\":\"2.0\",\"method\":\"theta.GetBlockByHeight\"," +
		"\"params\":[" +
		"{\"height\":\"" + fmt.Sprint(height) + "\"}]," +
		"\"id\":1}")
	if err != nil {
		return nil, err
	}

	result := theta.BlockResult_{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	if result.Error != nil && result.Error.Code != 0 {
		return nil, errors.New(result.Error.Message)
	}

	return &theta.BlockHeader{
		ChainID:          result.Result_.ChainID,
		Epoch:            result.Result_.Epoch,
		Height:           result.Result_.Height,
		Parent:           result.Result_.Parent,
		TransactionsHash: result.Result_.TransactionsHash,
		StateHash:        result.Result_.StateHash,
		Timestamp:        result.Result_.Timestamp,
		Proposer:         result.Result_.Proposer,
		Children:         result.Result_.Children,
		Status:           result.Result_.Status,
		Hash:             result.Result_.Hash,
		Hcc:              result.Result_.Hcc,
	}, nil
}

func (r *RpcHandler) GetBlock(ctx context.Context, hash string) (*theta.Block, error) {
	body, err := r.SendRpc("{\"jsonrpc\":\"2.0\",\"method\":\"theta.GetBlock\"," +
		"\"params\":[" +
		"{\"hash\":\"" + hash + "\"}]," +
		"\"id\":1}")
	if err != nil {
		return nil, err
	}

	result := theta.BlockResult_{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	if result.Error != nil && result.Error.Code != 0 {
		return nil, errors.New(result.Error.Message)
	}

	return result.Result_, nil
}

func (r *RpcHandler) GetBlockByHeight(ctx context.Context, height int64) (*theta.Block, error) {
	body, err := r.SendRpc("{\"jsonrpc\":\"2.0\",\"method\":\"theta.GetBlockByHeight\"," +
		"\"params\":[" +
		"{\"height\":\"" + fmt.Sprint(height) + "\"}]," +
		"\"id\":1}")
	if err != nil {
		return nil, err
	}

	result := theta.BlockResult_{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	if result.Error != nil && result.Error.Code != 0 {
		return nil, errors.New(result.Error.Message)
	}

	return result.Result_, nil
}

func (r *RpcHandler) GetAccount(ctx context.Context, address string) (*theta.Account, error) {
	body, err := r.SendRpc("{\n" +
		"\"jsonrpc\":\"2.0\",\n" +
		"\"method\":\"theta.GetAccount\",\n" +
		"\"params\":[{\"address\":\"" + address +"\"}],\n" +
		"\"id\":1\n}")
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

func (r *RpcHandler)SendToken(context context.Context, send *theta.SendToken) (*theta.BroadcastRawTransactionAsync, error) {
	prefix := "a9059cbb"
	decoded, err := hex.DecodeString(send.PrivateKey)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	privateKey, err := crypto.PrivateKeyFromBytes(decoded)
	if err != nil {
		log.Println("thriftrpcserver/handler.go:57", err)
		log.Println(len(decoded))
		return nil, err
	}

	if len(send.To) == 0 {
		return nil, fmt.Errorf("The from and to address cannot be empty")
	}
	if privateKey.PublicKey().Address().String() == send.To {
		return nil, fmt.Errorf("The from and to address cannot be identical")
	}

	data := prefix + "000000000000000000000000" + send.To + fmt.Sprintf("%640X", send.Amount)

	dataHex, err := hex.DecodeString(data)
	if err != nil {
		log.Println("thriftrpcserver/handler.go:71 ", err)
		return nil, err
	}

	acc, err := r.GetAccount(context, privateKey.PublicKey().Address().String())
	if err != nil {
		log.Println("thriftrpcserver/handler.go:77 ", err)
		return nil, err
	}

	seq, err := strconv.ParseInt(acc.Sequence, 10, 64)
	if err != nil {
		log.Println("thriftrpcserver/handler.go:83", err)
		return nil, err
	}

	from := types.TxInput{
		Address: privateKey.PublicKey().Address(),
		Coins: types.Coins{
			ThetaWei: new(big.Int).SetUint64(0),
			TFuelWei: new(big.Int).SetUint64(0),
		},
		Sequence: uint64(seq + 1),
	}

	to := types.TxOutput{
		Address: common.HexToAddress(send.To),
	}

	smartContractTx := &types.SmartContractTx{
		From:     from,
		To:       to,
		GasLimit: 10000000,
		GasPrice: big.NewInt(100000000),
		Data:     dataHex,
	}
	
	sig, err := privateKey.Sign(smartContractTx.SignBytes(r.chainID))
	if err != nil {
		return nil, err
	}

	smartContractTx.SetSignature(privateKey.PublicKey().Address(), sig)
	raw, err := types.TxToBytes(smartContractTx)
	if err != nil {
		utils.Error("Failed to encode transaction: %v\n", err)
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

	return result.Result_, nil
}

func (r *RpcHandler) GetTokenBalance(context context.Context, address, contractAddress, fromAddress string) (int64, error) {
	prefix := "70a08231"
	data := prefix + "000000000000000000000000" + address
	log.Println(data)

	acc, err := r.GetAccount(context, fromAddress)
	if err != nil {
		log.Println("thriftrpcserver/handler.go:74 ", err)
		return 0, err
	}

	seq, err := strconv.ParseInt(acc.Sequence, 10, 64)
	if err != nil {
		log.Println("thriftrpcserver/handler.go:80", err)
		return 0, err
	}

	from := types.TxInput{
		Address: common.HexToAddress(fromAddress),
		Coins: types.Coins{
			ThetaWei: new(big.Int).SetUint64(0),
			TFuelWei: new(big.Int).SetUint64(0),
		},
		Sequence: uint64(seq + 1),
	}

	to := types.TxOutput{
		Address: common.HexToAddress(contractAddress),
	}

	dataHex, err := hex.DecodeString(data)
	if err != nil {
		log.Println("thriftrpcserver/handler.go:87 ", err)
		return 0, err
	}

	smartContractTx := &types.SmartContractTx{
		From:     from,
		To:       to,
		GasLimit: 10000000,
		GasPrice: big.NewInt(100000000),
		Data:     dataHex,
	}

	raw, err := types.TxToBytes(smartContractTx)
	if err != nil {
		log.Println("thriftrpcserver/handler.go:116 ", err)
		return 0, err
	}

	signedTx := hex.EncodeToString(raw)
	msg := "{" +
		"\"jsonrpc\":\"2.0\"," +
		"\"method\":\"theta.CallSmartContract\"," +
		"\"params\":[" +
		"{\"sctx_bytes\":\""+ signedTx + "\"}]" +
		",\"id\":1}"
	body, err := r.SendRpc(msg)
	if err != nil {
		log.Println("thriftrpcserver/handler.go:129", err)
		return 0, err
	}

	result := theta.SmartContractCall{}
	err = json.Unmarshal(body, &result)

	balance, _ := strconv.ParseInt(result.VMReturn, 10, 64)

	return balance, nil
}

func (r *RpcHandler) SendTx(context context.Context, send *theta.Send) (*theta.BroadcastRawTransactionAsync, error) {
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

	if len(send.To) == 0 {
		return nil, fmt.Errorf("The from and to address cannot be empty")
	}
	if privateKey.PublicKey().Address().String() == send.To {
		return nil, fmt.Errorf("The from and to address cannot be identical")
	}

	chainID := r.chainID
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

	acc, err := r.GetAccount(context, privateKey.PublicKey().Address().String())
	if err != nil {
		log.Println(err)
		return nil, err
	}

	seq, err := strconv.ParseInt(acc.Sequence, 10, 64)
	if err != nil {
		log.Println("thriftrpcserver/handler.go:80", err)
		return nil, err
	}

	inputs := []types.TxInput{{
		Address: privateKey.PublicKey().Address(),
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

	signBytes := sendTx.SignBytes(chainID)
	key := &keystore.Key{
		Id:         uuid.NewRandom(),
		Address:    privateKey.PublicKey().Address(),
		PrivateKey: privateKey,
	}

	sig, err := key.Sign(signBytes)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	sendTx.SetSignature(privateKey.PublicKey().Address(), sig)
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
	re := map[string]interface{}{}
	_ = json.Unmarshal(body, &re)
	b, _ := json.MarshalIndent(re, "", "    ")

	log.Println(string(b))
	return body, nil
}
