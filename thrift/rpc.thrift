namespace go rpc.theta

enum TErrorCode{
    EGood = 0,
    ENotFound = -1,
    EUnknown = -2 ,
    EDataExisted = -3
}

struct Version {
	1: string version,
    2: string git_hash,
    3: string timestamp
}

struct Account {
	1: string sequence,
	2: Coin coin,
	3: list<string> reserverd_funds,
	4: string last_update_block_height,
	5: string root,
	6: string code
}

struct Proposer {
	1: string address,
	2: Coin coins,
	3: string sequence,
	4: string signature
}

struct Output {
	1: string address,
	2: Coin coins
}

struct Input {
	1: string address,
	2: Coin coins,
	3: string sequence,
	4: string signature
}

struct RawTransaction {
	1: Proposer proposer,
	2: list<Output> outputs,
	3: string block_height,
	4: Fee fee,
	5: list<Input> inputs,
}

struct TransactionInBlock {
	1: RawTransaction raws,
	2: i32 type,
	3: string hash
}

struct TransactionResult {
	1: string block_hash,
	2: string block_height,
	3: string status,
	4: string hash,
	5: TransactionData transaction
}

struct TransactionData {
	1: Fee fee,
	2: Input inputs,
	3: Output outputs
}

struct PendingTransaction {
	1: list<string> tx_hashes
}


struct BroadcastRawTransaction {
	1: string hash,
	2: TransactionBlock block
}

struct CallSmartContract {
	1: string result
}


struct TransactionBlock {
	1: string ChainID,
	2: i64 Epoch,
	3: i64 Height,
	4: string Parent,
	5: HCC HCC,
	6: string TxHash,
	7: string ReceiptHash,
	8: string Bloom,
	9: string StateHash,
	10: i64 Timestamp,
	11: string Proposer,
	12: string Signature
}

struct HCC {
	1: list<string> Votes,
	2: string BlockHash
}

struct BroadcastRawTransactionAsync {
	1: string hash
}

struct Block {
	1: string chain_id,
	2: string epoch,
	3: string height,
	4: string parent,
	5: string transactions_hash,
	6: string state_hash,
	7: string timestamp,
	8: string proposer,
	9: list<string> children,
	10: i32 status,
	11: string hash,
	12: list<TransactionBlock> transactions
}

struct Coin {
	1: string thetawei,
	2: string tfuelwei
}

struct Fee {
	1: string thetawei,
	2: string tfuelwei
}

struct SmartContract {
	1: string chain,
	2: string from_address,
	3: string gas_price,
	4: string gas_limit,
	5: string data,
	6: string seq,
	7: string to_address
}

struct NewKey {
	1: string address
}

struct ListKeys {
	1: list<string> addresses
}

struct StatusKey {
	1: bool unlocked
}

struct Send {
	1: string chain_id,
	2: string from_address,
	3: string to,
	4: string thetawei,
	5: string tfuelwei,
	6: string fee,
	7: string private_key
}

struct AccountResult {
    1: string jsonrpc,
    2: string id,
    3: Account result
}

struct BroadcastRawTransactionAsyncResult {
    1: string jsonrpc,
    2: string id,
    3: BroadcastRawTransactionAsync result
}

service ThetaService{
    Account getAccount(1: string account),
	BroadcastRawTransactionAsync sendTx(1: Send send)
}