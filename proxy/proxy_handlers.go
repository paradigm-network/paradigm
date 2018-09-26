package proxy

import (
	"bytes"
	"encoding/json"
	"github.com/paradigm-network/paradigm/accounts"
	"github.com/paradigm-network/paradigm/accounts/keystore"
	"github.com/paradigm-network/paradigm/common"
	"github.com/paradigm-network/paradigm/common/hexutil"
	"github.com/paradigm-network/paradigm/common/rlp"
	"github.com/paradigm-network/paradigm/types"
	"github.com/rs/zerolog/log"
	"io/ioutil"
	"math/big"
	"net/http"
)

/*
GET /account/{address}
example: /account/0x50bd8a037442af4cdf631495bcaa5443de19685d
returns: JSON JsonAccount
*/
func accountHandler(w http.ResponseWriter, r *http.Request, m *Service) {
	param := r.URL.Path[len("/account/"):]
	log.Info().Str("param", param).Msg("GET account")
	address := common.HexToAddress(param)
	log.Info().Str("address", address.Hex()).Msg("GET account")

	balance := m.state.GetBalance(address)
	nonce := m.state.GetNonce(address)
	account := JsonAccount{
		Address: address.Hex(),
		Balance: balance,
		Nonce:   nonce,
	}

	js, err := json.Marshal(account)
	if err != nil {
		log.Error().Err(err).Msg("Marshaling JSON response")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

/*
GET /accounts
returns: JSON JsonAccountList
*/
func accountsHandler(w http.ResponseWriter, r *http.Request, m *Service) {
	log.Info().Msg("GET accounts")

	var al JsonAccountList

	for _, account := range m.keyStore.Accounts() {
		balance := m.state.GetBalance(account.Address)
		nonce := m.state.GetNonce(account.Address)
		al.Accounts = append(al.Accounts,
			JsonAccount{
				Address: account.Address.Hex(),
				Balance: balance,
				Nonce:   nonce,
			})
	}

	js, err := json.Marshal(al)
	if err != nil {
		log.Error().Err(err).Msg("Marshaling JSON response")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}
/*
POST /tx
data: JSON SendTxArgs
returns: JSON JsonTxRes

This endpoints allows calling SmartContract code for NON-READONLY operations.
These operations can MODIFY the EVM state.

The data does NOT need to be SIGNED. In fact, this endpoint is meant to be used
for transactions whose originator is an account CONTROLLED by the 
Service (ie. present in the Keystore).

The Nonce field is not necessary either since the Service will fetch it from the
State.

This is an ASYNCHRONOUS operation. It will return the hash of the transaction that
was SUBMITTED to  but there is no guarantee that the transactions will
get applied to the State.

One should use the /receipt endpoint to retrieve the corresponding receipt and
verify if/how the State was modified.
*/
func transactionHandler(w http.ResponseWriter, r *http.Request, m *Service) {
	decoder := json.NewDecoder(r.Body)
	var txArgs SendTxArgs
	err := decoder.Decode(&txArgs)
	if err != nil {
		log.Error().Err(err).Msg("Decoding JSON txArgs")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()
	log.Info().Interface("txArgs", txArgs).Msg("POST tx")
	tx, err := prepareTransaction(txArgs, m.state, m.keyStore)
	if err != nil {
		log.Error().Err(err).Msg("Preparing Transaction")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data, err := rlp.EncodeToBytes(tx)
	if err != nil {
		log.Error().Err(err).Msg("Encoding Transaction")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Info().Msg("submitting tx")
	m.submitCh <- data
	log.Info().Msg("submitted tx")

	res := JsonTxRes{TxHash: tx.Hash().Hex()}
	js, err := json.Marshal(res)
	if err != nil {
		log.Error().Err(err).Msg("Marshalling JSON response")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)

}

/*
POST /rawtx
data: STRING Hex representation of the raw transaction bytes
	  ex: 0xf8620180830f4240946266b0dd0116416b1dacf36...
returns: JSON JsonTxRes

This endpoint allows sending NON-READONLY transactions ALREADY SIGNED. The client
is left to compose a transaction, sign it and RLP encode it. The resulting bytes,
represented as a Hex string is passed to this method to be forwarded to the EVM.

This allows executing transactions on behalf of accounts that are NOT CONTROLLED
by the  service.

Like the /tx endpoint, this is an ASYNCHRONOUS operation and the effect on the
State should be verified by fetching the transaction' receipt.
*/
func rawTransactionHandler(w http.ResponseWriter, r *http.Request, m *Service) {
	log.Info().Interface("request", r).Msg("POST rawtx")

	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Error().Err(err).Msg("Reading request body")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sBody := string(body)
	log.Info().Str("body", string(body))
	rawTxBytes, err := hexutil.Decode(sBody)
	if err != nil {
		log.Error().Err(err).Msg("Reading raw tx from request body")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Info().Bytes("raw tx bytes", rawTxBytes)

	log.Info().Msg("submitting tx")
	m.submitCh <- rawTxBytes
	log.Info().Msg("submitted tx")

	var t types.Transaction
	if err := rlp.Decode(bytes.NewReader(rawTxBytes), &t); err != nil {
		log.Error().Err(err).Msg("Decoding Transaction")
		return
	}
	log.Info().Str("hash", t.Hash().Hex()).Msg("Decoded tx")

	res := JsonTxRes{TxHash: t.Hash().Hex()}
	js, err := json.Marshal(res)
	if err != nil {
		log.Error().Err(err).Msg("Marshalling JSON response")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)

}

/*
GET /tx/{tx_hash}
ex: /tx/0xbfe1aa80eb704d6342c553ac9f423024f448f7c74b3e38559429d4b7c98ffb99
returns: JSON JsonReceipt
*/
func transactionReceiptHandler(w http.ResponseWriter, r *http.Request, m *Service) {
	param := r.URL.Path[len("/tx/"):]
	txHash := common.HexToHash(param)
	log.Info().Str("tx_hash", txHash.Hex()).Msg("GET tx")

	tx, err := m.state.GetTransaction(txHash)
	if err != nil {
		log.Error().Err(err).Msg("Getting Transaction")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	receipt, err := m.state.GetReceipt(txHash)
	if err != nil {
		log.Error().Err(err).Msg("Getting Receipt")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	signer := types.NewBasicSigner()
	from, err := types.Sender(signer, tx)
	if err != nil {
		log.Error().Err(err).Msg("Getting Tx Sender")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jsonReceipt := JsonReceipt{
		Root:              common.BytesToHash(receipt.PostState),
		TransactionHash:   txHash,
		From:              from,
		To:                tx.To(),
		GasUsed:           receipt.GasUsed,
		CumulativeGasUsed: receipt.CumulativeGasUsed,
		ContractAddress:   receipt.ContractAddress,
		Logs:              receipt.Logs,
		LogsBloom:         receipt.Bloom,
	}

	if receipt.Logs == nil {
		jsonReceipt.Logs = []*types.Log{}
	}

	js, err := json.Marshal(jsonReceipt)
	if err != nil {
		log.Error().Err(err).Msg("Marshaling JSON response")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func prepareTransaction(args SendTxArgs, state *State, ks *keystore.KeyStore) (*types.Transaction, error) {
	var err error
	args, err = prepareSendTxArgs(args)
	if err != nil {
		return nil, err
	}

	if args.Nonce == nil {
		args.Nonce = new(uint64)
		*args.Nonce = state.GetNonce(args.From)
	}

	var tx *types.Transaction
	if args.To == nil {
		tx = types.NewContractCreation(*args.Nonce,
			args.Value,
			args.Gas,
			args.GasPrice,
			common.FromHex(args.Data))
	} else {
		tx = types.NewTransaction(*args.Nonce,
			*args.To,
			args.Value,
			args.Gas,
			args.GasPrice,
			common.FromHex(args.Data))
	}

	signer := types.NewBasicSigner()

	account, err := ks.Find(accounts.Account{Address: args.From})
	if err != nil {
		return nil, err
	}
	signature, err := ks.SignHash(account, signer.Hash(tx).Bytes())
	if err != nil {
		return nil, err
	}
	signedTx, err := tx.WithSignature(signer, signature)
	if err != nil {
		return nil, err
	}

	return signedTx, nil
}

func prepareSendTxArgs(args SendTxArgs) (SendTxArgs, error) {
	if args.Gas == nil {
		args.Gas = defaultGas
	}
	if args.GasPrice == nil {
		args.GasPrice = big.NewInt(0)
	}
	if args.Value == nil {
		args.Value = big.NewInt(0)
	}
	return args, nil
}


type JsonAccount struct {
	Address string   `json:"address"`
	Balance *big.Int `json:"balance"`
	Nonce   uint64   `json:"nonce"`
}

type JsonAccountList struct {
	Accounts []JsonAccount `json:"accounts"`
}

// SendTxArgs represents the arguments to sumbit a new transaction into the transaction pool.
type SendTxArgs struct {
	From     common.Address  `json:"from"`
	To       *common.Address `json:"to"`
	Gas      *big.Int        `json:"gas"`
	GasPrice *big.Int        `json:"gasPrice"`
	Value    *big.Int        `json:"value"`
	Data     string          `json:"data"`
	Nonce    *uint64         `json:"nonce"`
}

type JsonCallRes struct {
	Data string `json:"data"`
}

type JsonTxRes struct {
	TxHash string `json:"txHash"`
}

type JsonReceipt struct {
	Root              common.Hash     `json:"root"`
	TransactionHash   common.Hash     `json:"transactionHash"`
	From              common.Address  `json:"from"`
	To                *common.Address `json:"to"`
	GasUsed           *big.Int        `json:"gasUsed"`
	CumulativeGasUsed *big.Int        `json:"cumulativeGasUsed"`
	ContractAddress   common.Address  `json:"contractAddress"`
	Logs              []*types.Log `json:"logs"`
	LogsBloom         types.Bloom  `json:"logsBloom"`

}
