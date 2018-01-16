package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"

	"github.com/tendermint/abci/client"

	"net/http" // Provides HTTP client and server implementations.

	"github.com/blockfreight/go-bftx/lib/app/bf_tx"
	"github.com/blockfreight/go-bftx/lib/pkg/crypto"
	"github.com/blockfreight/go-bftx/lib/pkg/leveldb"
	rpc "github.com/tendermint/tendermint/rpc/client"
	tmTypes "github.com/tendermint/tendermint/types"

	// Provides HTTP client and server implementations.
	// ===============
	// Tendermint Core
	// ===============
	abciTypes "github.com/tendermint/abci/types"
)

var TendermintClient abcicli.Client

func ConstructBfTx(transaction bf_tx.BF_TX) (interface{}, error) {

	resInfo, err := TendermintClient.InfoSync(abciTypes.RequestInfo{})
	if err != nil {
		return nil, errors.New(strconv.Itoa(http.StatusInternalServerError))
	}

	hash, err := bf_tx.HashBFTX(transaction)
	if err != nil {
		return nil, errors.New(strconv.Itoa(http.StatusInternalServerError))
	}

	// Generate BF_TX id
	transaction.Id = bf_tx.GenerateBFTXUID(hash, resInfo.LastBlockAppHash)

	/*jsonContent, err := json.Marshal(transaction)
	if err != nil {
		return nil, errors.New(strconv.Itoa(http.StatusInternalServerError))
	}

	transaction.Private = string(crypto.CryptoTransaction(string(jsonContent)))*/

	// Get the BF_TX content in string format
	content, err := bf_tx.BFTXContent(transaction)
	if err != nil {
		return nil, errors.New(strconv.Itoa(http.StatusInternalServerError))
	}

	/* TODO: ENCRYPT TRANSACTION */

	// Save on DB
	if err = leveldb.RecordOnDB(transaction.Id, content); err != nil {
		return nil, errors.New(strconv.Itoa(http.StatusInternalServerError))
	}

	return transaction, nil
}

func SignBfTx(idBftx string) (interface{}, error) {
	transaction, err := leveldb.GetBfTx(idBftx)

	if err != nil {
		if err.Error() == "LevelDB Get function: BF_TX not found." {
			return nil, errors.New(strconv.Itoa(http.StatusNotFound))
		}
		return nil, errors.New(strconv.Itoa(http.StatusInternalServerError))
	}

	if transaction.Verified {
		return nil, errors.New(strconv.Itoa(http.StatusNotAcceptable))
	}

	// Sign BF_TX
	transaction, err = crypto.SignBFTX(transaction)
	if err != nil {
		return nil, errors.New(strconv.Itoa(http.StatusInternalServerError))
	}

	/*jsonContent, err := json.Marshal(transaction)
	if err != nil {
		return nil, errors.New(strconv.Itoa(http.StatusInternalServerError))
	}

	transaction.Private = string(crypto.CryptoTransaction(string(jsonContent)))*/

	// Get the BF_TX content in string format
	content, err := bf_tx.BFTXContent(transaction)
	if err != nil {
		return nil, errors.New(strconv.Itoa(http.StatusInternalServerError))
	}

	// Update on DB
	if err = leveldb.RecordOnDB(string(transaction.Id), content); err != nil {
		return nil, errors.New(strconv.Itoa(http.StatusInternalServerError))
	}

	return transaction, nil
}

func BroadcastBfTx(idBftx string) (interface{}, error) {
	rpcClient := rpc.NewHTTP("tcp://127.0.0.1:46657", "/websocket")
	err := rpcClient.Start()
	if err != nil {
		fmt.Println("Error when initializing rpcClient")
		log.Fatal(err.Error())
	}

	defer rpcClient.Stop()

	// Get a BF_TX by id
	transaction, err := leveldb.GetBfTx(idBftx)
	if err != nil {
		if err.Error() == "LevelDB Get function: BF_TX not found." {
			return nil, errors.New(strconv.Itoa(http.StatusNotFound))
		}
		return nil, errors.New(strconv.Itoa(http.StatusInternalServerError))
	}

	if !transaction.Verified {
		return nil, errors.New(strconv.Itoa(http.StatusNotAcceptable))
	}
	if transaction.Transmitted {
		return nil, errors.New(strconv.Itoa(http.StatusNotAcceptable))
	}

	/* TODO: ENCRYPT TRANSACTION */

	// Change the boolean valud for Transmitted attribute
	transaction.Transmitted = true

	// Get the BF_TX content in string format
	content, err := bf_tx.BFTXContent(transaction)
	if err != nil {
		return nil, errors.New(strconv.Itoa(http.StatusInternalServerError))
	}

	// Update on DB
	if err = leveldb.RecordOnDB(string(transaction.Id), content); err != nil {
		return nil, errors.New(strconv.Itoa(http.StatusInternalServerError))
	}

	var tx tmTypes.Tx
	tx = []byte(content)

	_, rpcErr := rpcClient.BroadcastTxSync(tx)
	if rpcErr != nil {
		fmt.Printf("%+v\n", rpcErr)
		return nil, rpcErr
	}

	return transaction, nil
}

func GetTransaction(idBftx string) (interface{}, error) {
	transaction, err := leveldb.GetBfTx(idBftx)
	if err != nil {
		if err.Error() == "LevelDB Get function: BF_TX not found." {
			return nil, errors.New(strconv.Itoa(http.StatusNotFound))
		}
		return nil, errors.New(strconv.Itoa(http.StatusInternalServerError))
	}

	/* TODO: DECRYPT TRANSACTION */

	return transaction, nil
}

func QueryTransaction(idBftx string) (interface{}, error) {
	rpcClient := rpc.NewHTTP("tcp://127.0.0.1:46657", "/websocket")
	err := rpcClient.Start()
	if err != nil {
		fmt.Println("Error when initializing rpcClient")
		log.Fatal(err.Error())
	}
	defer rpcClient.Stop()
	query := "bftx.id='" + idBftx + "'"
	resQuery, err := rpcClient.TxSearch(query, true)
	if err != nil {
		return nil, err
	}

	if len(resQuery) > 0 {
		var transaction bf_tx.BF_TX
		err := json.Unmarshal(resQuery[0].Tx, &transaction)
		if err != nil {
			return nil, err
		}

		return transaction, nil
	}

	return nil, errors.New(strconv.Itoa(http.StatusNotFound))
}
