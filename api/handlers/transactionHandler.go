package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"os"
	"reflect"
	"runtime"
	"strconv"
	"time"

	"net/http" // Provides HTTP client and server implementations.

	"github.com/blockfreight/go-bftx/lib/app/bf_tx"
	"github.com/blockfreight/go-bftx/lib/pkg/common"
	"github.com/blockfreight/go-bftx/lib/pkg/leveldb"
	"github.com/blockfreight/go-bftx/lib/pkg/saberservice"
	// Provides HTTP client and server implementations.
	// ===============
	// Tendermint Core
	// ===============
)

var TendermintClient abcicli.Client

func getFunctionName(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}

// simpleLogger writes errors and the function name that generated the error to bftx.log
func simpleLogger(i interface{}, currentError error) {
	// If the file doesn't exist, create it, or append to the file
	f, err := os.OpenFile(os.Getenv("GOPATH")+"/src/github.com/blockfreight/go-bftx/logs/api.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	if _, err := f.Write([]byte(time.Now().Format("2006/01/02 15:04") + ", " + getFunctionName(i) + ", " + currentError.Error() + "\n\n")); err != nil {
		log.Fatal(err)
	}
	if err := f.Close(); err != nil {
		log.Fatal(err)
	}
}

// queryLogger writes errors, the function name that generated the error, and the transaction body to bftx.log for cmdQuery only
func queryLogger(i interface{}, currentError string, id string) {
	// If the file doesn't exist, create it, or append to the file
	f, err := os.OpenFile(os.Getenv("GOPATH")+"/src/github.com/blockfreight/go-bftx/logs/bftx.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	if _, err := f.Write([]byte(time.Now().Format("2006/01/02 15:04") + ", " + getFunctionName(i) + ", " + currentError + ", " + id + "\n\n")); err != nil {
		log.Fatal(err)
	}
	if err := f.Close(); err != nil {
		log.Fatal(err)
	}
}

// transLogger writes errors, the function name that generated the error, and the transaction body to bftx.log
func transLogger(i interface{}, currentError error, id string) {
	// If the file doesn't exist, create it, or append to the file
	f, err := os.OpenFile(os.Getenv("GOPATH")+"/src/github.com/blockfreight/go-bftx/logs/api.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	if _, err := f.Write([]byte(time.Now().Format("2006/01/02 15:04") + ", " + getFunctionName(i) + ", " + currentError.Error() + ", " + id + "\n\n")); err != nil {
		log.Fatal(err)
	}
	if err := f.Close(); err != nil {
		log.Fatal(err)
	}
}

func ConstructBfTx(transaction bf_tx.BF_TX) (interface{}, error) {
	if err := transaction.GenerateBFTXUID(common.ORIGIN_API); err != nil {
		return nil, err
	}

	return transaction, nil
}

func SignBfTx(idBftx string) (interface{}, error) {
	var transaction bf_tx.BF_TX
	if err := transaction.SignBFTX(idBftx, common.ORIGIN_API); err != nil {
		return nil, err
	}

	return transaction, nil
}

func EncryptBfTx(idBftx string) (interface{}, error) {
	var transaction bf_tx.BF_TX
	data, err := leveldb.GetBfTx(idBftx)
	if err != nil {
		if err.Error() == "LevelDB Get function: BF_TX not found." {
			transLogger(EncryptBfTx, err, idBftx)
			return nil, errors.New(strconv.Itoa(http.StatusNotFound))
		}
		transLogger(EncryptBfTx, err, idBftx)
		return nil, errors.New(strconv.Itoa(http.StatusInternalServerError))
	}

	json.Unmarshal(data, &transaction)
	if err != nil {
		transLogger(SignBfTx, err, idBftx)
		return nil, errors.New(strconv.Itoa(http.StatusInternalServerError))
	}

	if transaction.Verified {
		return nil, errors.New(strconv.Itoa(http.StatusNotAcceptable))
	}

	nwbftx, err := saberservice.BftxStructConverstionON(&transaction)
	if err != nil {
		log.Fatalf("Conversion error, can not convert old bftx to new bftx structure")
		transLogger(EncryptBfTx, err, idBftx)
		return nil, errors.New(strconv.Itoa(http.StatusInternalServerError))
	}
	st := saberservice.SaberDefaultInput()
	saberbftx, err := saberservice.SaberEncoding(nwbftx, st)
	if err != nil {
		transLogger(EncryptBfTx, err, idBftx)
		return nil, errors.New(strconv.Itoa(http.StatusInternalServerError))
	}
	bftxold, err := saberservice.BftxStructConverstionNO(saberbftx)
	//update the encoded transaction to database
	// Get the BF_TX content in string format
	content, err := bf_tx.BFTXContent(*bftxold)
	if err != nil {
		transLogger(EncryptBfTx, err, idBftx)
		return nil, errors.New(strconv.Itoa(http.StatusInternalServerError))
	}

	// Update on DB
	err = leveldb.RecordOnDB(string(bftxold.Id), content)
	if err != nil {
		transLogger(EncryptBfTx, err, idBftx)
		return nil, errors.New(strconv.Itoa(http.StatusInternalServerError))
	}

	return bftxold, nil
}

func DecryptBfTx(idBftx string) (interface{}, error) {
	var transaction bf_tx.BF_TX
	data, err := leveldb.GetBfTx(idBftx)
	if err != nil {
		return nil, errors.New(strconv.Itoa(http.StatusInternalServerError))
	}

	json.Unmarshal(data, &transaction)

	if err != nil {
		if err.Error() == "LevelDB Get function: BF_TX not found." {
			transLogger(DecryptBfTx, err, idBftx)
			return nil, errors.New(strconv.Itoa(http.StatusNotFound))
		}
		transLogger(DecryptBfTx, err, idBftx)
		return nil, errors.New(strconv.Itoa(http.StatusInternalServerError))
	}

	if transaction.Verified {
		return nil, errors.New(strconv.Itoa(http.StatusNotAcceptable))
	}

	nwbftx, err := saberservice.BftxStructConverstionON(&transaction)
	if err != nil {
		log.Fatalf("Conversion error, can not convert old bftx to new bftx structure")
		transLogger(DecryptBfTx, err, idBftx)
		return nil, errors.New(strconv.Itoa(http.StatusInternalServerError))
	}
	st := saberservice.SaberDefaultInput()
	saberbftx, err := saberservice.SaberDecoding(nwbftx, st)
	if err != nil {
		transLogger(DecryptBfTx, err, idBftx)
		return nil, errors.New(strconv.Itoa(http.StatusInternalServerError))
	}
	bftxold, err := saberservice.BftxStructConverstionNO(saberbftx)
	//update the encoded transaction to database
	// Get the BF_TX content in string format
	content, err := bf_tx.BFTXContent(*bftxold)
	if err != nil {
		transLogger(DecryptBfTx, err, idBftx)
		return nil, errors.New(strconv.Itoa(http.StatusInternalServerError))
	}

	// Update on DB
	err = leveldb.RecordOnDB(string(bftxold.Id), content)
	if err != nil {
		transLogger(DecryptBfTx, err, idBftx)
		return nil, errors.New(strconv.Itoa(http.StatusInternalServerError))
	}

	return bftxold, nil
}

func BroadcastBfTx(idBftx string) (interface{}, error) {
	var transaction bf_tx.BF_TX

	if err := transaction.BroadcastBFTX(idBftx, common.ORIGIN_API); err != nil {
		return nil, err
	}

	return transaction, nil
}

func GetTotal() (interface{}, error) {
	var bftx bf_tx.BF_TX

	total, err := bftx.GetTotal()
	if err != nil {
		simpleLogger(GetTotal, err)
		return nil, err
	}

	return total, nil
}

func GetTransaction(idBftx string) (interface{}, error) {
	var transaction bf_tx.BF_TX
	if err := transaction.GetBFTX(idBftx, common.ORIGIN_API); err != nil {
		return nil, err
	}

	return transaction, nil
}

func QueryTransaction(idBftx string) (interface{}, error) {
	var transaction bf_tx.BF_TX
	if err := transaction.QueryBFTX(idBftx, common.ORIGIN_API); err != nil {
		return nil, err
	}

	return transaction, nil
}
