package rpc

import (
	"encoding/hex"
	"errors"
	"sync"
	"time"

	"theta/common"
	"theta/common/hexutil"
	"theta/core"
	"theta/crypto"
)

const txTimeout = 60 * time.Second

type Callback struct {
	txHash   string
	created  time.Time
	Callback func(*core.Block)
}

type TxCallbackManager struct {
	mu               *sync.Mutex
	txHashToCallback map[string]*Callback
	callbacks        []*Callback
}

func NewTxCallbackManager() *TxCallbackManager {
	return &TxCallbackManager{
		mu:               &sync.Mutex{},
		txHashToCallback: make(map[string]*Callback),
		callbacks:        []*Callback{},
	}
}

func (m *TxCallbackManager) AddCallback(txHash common.Hash, cb func(*core.Block)) {
	m.mu.Lock()
	defer m.mu.Unlock()

	txHashStr := txHash.Hex()
	callback := &Callback{
		txHash:   txHashStr,
		created:  time.Now(),
		Callback: cb,
	}
	_, exists := m.txHashToCallback[txHashStr]
	if exists {
		logger.Infof("Overwriting tx callback, txHash=%v", txHashStr)
	}
	m.txHashToCallback[txHashStr] = callback
	m.callbacks = append(m.callbacks, callback)
}

func (m *TxCallbackManager) RemoveCallback(txHash common.Hash) (cb *Callback, exists bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := txHash.Hex()
	cb, exists = m.txHashToCallback[key]
	if exists {
		delete(m.txHashToCallback, key)
	}
	return
}

func (m *TxCallbackManager) Trim() {
	m.mu.Lock()
	defer m.mu.Unlock()

	i := 0
	for ; i < len(m.callbacks); i++ {
		cb := m.callbacks[i]
		if time.Since(cb.created) < txTimeout {
			break
		}
		cb2, ok := m.txHashToCallback[cb.txHash]
		if ok && cb2.created == cb.created {
			logger.Infof("Removing timedout tx callback, txHash=%v", cb.txHash)
			delete(m.txHashToCallback, cb.txHash)
		}
	}
	m.callbacks = m.callbacks[i:]
}

var txCallbackManager = NewTxCallbackManager()

func (t *ThetaRPCService) txCallback() {
	defer t.wg.Done()

	timer := time.NewTicker(1 * time.Second)
	defer timer.Stop()

	for {
		select {
		case <-t.ctx.Done():
			logger.Infof("ctx.Done()")
			return
		case block := <-t.consensus.FinalizedBlocks():
			logger.Infof("Processing finalized block, height=%v", block.Height)

			for _, tx := range block.Txs {
				txHash := crypto.Keccak256Hash(tx)
				cb, ok := txCallbackManager.RemoveCallback(txHash)
				if ok {
					go cb.Callback(block)
				}
			}

			logger.Infof("Done processing finalized block, height=%v", block.Height)
		case <-timer.C:
			logger.Debugf("txCallbackManager.Trim()")

			txCallbackManager.Trim()

			logger.Debugf("Done txCallbackManager.Trim()")
		}
	}
}

// ------------------------------- BroadcastRawTransaction -----------------------------------

type BroadcastRawTransactionArgs struct {
	TxBytes string `json:"tx_bytes"`
}

type BroadcastRawTransactionResult struct {
	TxHash string            `json:"hash"`
	Block  *core.BlockHeader `json:"block",rlp:"nil"`
}

func (t *ThetaRPCService) BroadcastRawTransaction(
	args *BroadcastRawTransactionArgs, result *BroadcastRawTransactionResult) (err error) {
	txBytes, err := decodeTxHexBytes(args.TxBytes)
	if err != nil {
		return err
	}

	hash := crypto.Keccak256Hash(txBytes)
	result.TxHash = hash.Hex()

	logger.Infof("Broadcast raw transaction (sync): %v, hash: %v", hex.EncodeToString(txBytes), hash.Hex())

	err = t.mempool.InsertTransaction(txBytes)
	if err != nil {
		return err
	}

	t.mempool.BroadcastTx(txBytes)

	finalized := make(chan *core.Block)
	timeout := time.NewTimer(txTimeout)
	defer timeout.Stop()

	txCallbackManager.AddCallback(hash, func(block *core.Block) {
		select {
		case finalized <- block:
		default:
		}
	})

	select {
	case block := <-finalized:
		if block == nil {
			logger.Infof("Tx callback returns nil, txHash=%v", result.TxHash)
			return errors.New("Internal server error")
		}
		result.Block = block.BlockHeader
		return nil
	case <-timeout.C:
		return errors.New("Timed out waiting for transaction to be included")
	}
}

// ------------------------------- BroadcastRawTransactionAsync -----------------------------------

type BroadcastRawTransactionAsyncArgs struct {
	TxBytes string `json:"tx_bytes"`
}

type BroadcastRawTransactionAsyncResult struct {
	TxHash string `json:"hash"`
}

func (t *ThetaRPCService) BroadcastRawTransactionAsync(
	args *BroadcastRawTransactionAsyncArgs, result *BroadcastRawTransactionAsyncResult) (err error) {
	txBytes, err := decodeTxHexBytes(args.TxBytes)
	if err != nil {
		return err
	}

	hash := crypto.Keccak256Hash(txBytes)
	result.TxHash = hash.Hex()

	logger.Infof("Broadcast raw transaction (async): %v, hash: %v", hex.EncodeToString(txBytes), hash.Hex())

	err = t.mempool.InsertTransaction(txBytes)
	if err != nil {
		return err
	}

	t.mempool.BroadcastTx(txBytes)

	return nil
}

// -------------------------- Utilities -------------------------- //

func decodeTxHexBytes(txBytes string) ([]byte, error) {
	if hexutil.Has0xPrefix(txBytes) {
		txBytes = txBytes[2:]
	}
	return hex.DecodeString(txBytes)
}
