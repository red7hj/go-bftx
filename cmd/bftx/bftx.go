// File: ./blockfreight/cmd/bftx/bftx.go
// Summary: Application code for Blockfreight™ | The blockchain of global freight.
// License: MIT License
// Company: Blockfreight, Inc.
// Author: Julian Nunez, Neil Tran, Julian Smith, Gian Felipe & contributors
// Site: https://blockfreight.com
// Support: <support@blockfreight.com>

// Copyright © 2017 Blockfreight, Inc. All Rights Reserved.

// Permission is hereby granted, free of charge, to any person obtaining
// a copy of this software and associated documentation files (the "Software"),
// to deal in the Software without restriction, including without limitation
// the rights to use, copy, modify, merge, publish, distribute, sublicense,
// and/or sell copies of the Software, and to permit persons to whom the
// Software is furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be
// included in all copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS
// OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
// WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
// CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

// =================================================================================================================================================
// =================================================================================================================================================
//
// BBBBBBBBBBBb     lll                                kkk             ffff                         iii                  hhh            ttt
// BBBB``````BBBB   lll                                kkk            fff                           ```                  hhh            ttt
// BBBB      BBBB   lll      oooooo        ccccccc     kkk    kkkk  fffffff  rrr  rrr    eeeee      iii     gggggg ggg   hhh  hhhhh   tttttttt
// BBBBBBBBBBBB     lll    ooo    oooo    ccc    ccc   kkk   kkk    fffffff  rrrrrrrr eee    eeee   iii   gggg   ggggg   hhhh   hhhh  tttttttt
// BBBBBBBBBBBBBB   lll   ooo      ooo   ccc           kkkkkkk        fff    rrrr    eeeeeeeeeeeee  iii  gggg      ggg   hhh     hhh    ttt
// BBBB       BBB   lll   ooo      ooo   ccc           kkkk kkkk      fff    rrr     eeeeeeeeeeeee  iii   ggg      ggg   hhh     hhh    ttt
// BBBB      BBBB   lll   oooo    oooo   cccc    ccc   kkk   kkkk     fff    rrr      eee      eee  iii    ggg    gggg   hhh     hhh    tttt    ....
// BBBBBBBBBBBBB    lll     oooooooo       ccccccc     kkk     kkkk   fff    rrr       eeeeeeeee    iii     gggggg ggg   hhh     hhh     ttttt  ....
//                                                                                                        ggg      ggg
//   Blockfreight™ | The blockchain of global freight.                                                      ggggggggg
//
// =================================================================================================================================================
// =================================================================================================================================================

// Initializes BFTX app to interacts with the Blockfreight™ Network.
package main

import (
	// =======================
	// Golang Standard library
	// =======================
	"bufio" // Implements buffered I/O.
	// Package csv reads and writes comma-separated values (CSV) files.
	"encoding/hex" // Implements hexadecimal encoding and decoding.
	"errors"       // Implements functions to manipulate errors.
	"fmt"          // Implements formatted I/O with functions analogous to C's printf and scanf.
	"io"           // Provides basic interfaces to I/O primitives.
	"log"          // Implements a simple logging package.
	"os"           // Provides a platform-independent interface to operating system functionality.
	"reflect"      // Implements run-time reflection, allowing a program to manipulate objects with arbitrary types.
	"runtime"      // Contains operations that interact with Go's runtime system.
	"strconv"      // Implements conversions to and from string representations of basic data types.
	"strings"      // Implements simple functions to manipulate UTF-8 encoded strings.
	"time"         // Provides functionality for measuring and displaying time.

	// ====================
	// Third-party packages
	// ====================

	"github.com/urfave/cli" // Provides structure and function to build command line apps in Go.

	// ===============
	// Tendermint Core
	// ===============
	"github.com/tendermint/abci/client"
	"github.com/tendermint/abci/types"
	rpc "github.com/tendermint/tendermint/rpc/client"
	tmTypes "github.com/tendermint/tendermint/types"

	// ======================
	// Blockfreight™ packages
	// ======================

	"github.com/blockfreight/go-bftx/api/handlers"
	"github.com/blockfreight/go-bftx/build/package/version" // Defines the current version of the project.
	"github.com/blockfreight/go-bftx/lib/app/bf_tx"         // Defines the Blockfreight™ Transaction (BF_TX) transaction standard and provides some useful functions to work with the BF_TX.
	"github.com/blockfreight/go-bftx/lib/app/validator"     // Provides functions to assure the input JSON is correct.
	"github.com/blockfreight/go-bftx/lib/pkg/crypto"        // Provides useful functions to sign BF_TX.
	"github.com/blockfreight/go-bftx/lib/pkg/leveldb"       // Provides some useful functions to work with LevelDB.
	"github.com/blockfreight/go-bftx/lib/pkg/saberservice"  // Provides function for saber-service.
)

// Structure for data passed to print response.
type response struct {
	// generic abci response
	Data   []byte
	Code   uint32
	Log    string
	Result string //Blockfreight Purposes

	Query *queryResponse
}

type queryResponse struct {
	Key    []byte
	Value  []byte
	Height int64
	Proof  []byte
}

// client is a global variable so it can be reused by the console
var client abcicli.Client
var rpcClient *rpc.HTTP

func main() {

	//workaround for the cli library (https://github.com/urfave/cli/issues/565)
	cli.OsExiter = func(_ int) {}

	app := cli.NewApp()
	app.Name = "bftx"
	app.Usage = "bftx [command] [args...]"
	app.Version = version.Version
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "address",
			Value: "tcp://127.0.0.1:46658",
			Usage: "address of application socket",
		},
		cli.StringFlag{
			Name:  "call",
			Value: "socket",
			Usage: "socket or grpc",
		},
		cli.BoolFlag{
			Name:  "verbose",
			Usage: "print the command and results as if it were a console session",
		},
		/*cli.StringFlag{
		    Name: "lang",
		    Value: "english",
		    Usage: "language for the greeting",
		},*/
		cli.StringFlag{
			Name:  "json_path, jp",
			Value: "./examples/",
			Usage: "define the source path where the json is",
		},
	}
	app.Commands = []cli.Command{
		{
			Name:  "batch",
			Usage: "Run a batch of Blockfreight™ commands against an application",
			Action: func(c *cli.Context) error {
				return cmdBatch(app, c)
			},
		},
		{
			Name:  "console",
			Usage: "Start an interactive Blockfreight™ console for multiple commands",
			Action: func(c *cli.Context) error {
				return cmdConsole(app, c)
			},
		},
		/*{
		    Name:  "echo",
		    Usage: "Have the application echo a message (Parameters: value_to_print)",
		    Action: func(c *cli.Context) error {
		        return cmdEcho(c)
		    },
		},*/
		{
			Name:  "info",
			Usage: "Get some info about the application (Parameters: none)",
			Action: func(c *cli.Context) error {
				return cmdInfo(c)
			},
		},
		{
			Name:  "set_option",
			Usage: "Set an option on the application (Parameters: --Global Options, value)",
			Action: func(c *cli.Context) error {
				return cmdSetOption(c)
			},
		},
		{
			Name:  "verify",
			Usage: "Verify the JSON imput against a BF_TX (Parameters: JSON Filepath)",
			Action: func(c *cli.Context) error {
				return cmdVerifyBfTx(c) //cmdCheckBfTx
			},
		},
		{
			Name:  "validate",
			Usage: "Validate a BF_TX (Parameters: JSON Filepath)",
			Action: func(c *cli.Context) error {
				return cmdValidateBfTx(c)
			},
		},
		{
			Name:  "construct",
			Usage: "Construct a new BF_TX (Parameters: JSON Filepath)",
			Action: func(c *cli.Context) error {
				return cmdConstructBfTx(c)
			},
		},
		{
			Name:  "sign",
			Usage: "Sign a new BF_TX (Parameters: BF_TX id)",
			Action: func(c *cli.Context) error {
				return cmdSignBfTx(c)
			},
		},
		{
			Name:  "broadcast",
			Usage: "Deliver a new BF_TX to application (Parameters: BF_TX id)",
			Action: func(c *cli.Context) error {
				return cmdBroadcastBfTx(c)
			},
		},
		{
			Name:  "commit",
			Usage: "Commit the application state and return the Merkle root hash (Parameters: none)",
			Action: func(c *cli.Context) error {
				return cmdCommit(c)
			},
		},
		{
			Name:  "query",
			Usage: "Query application state",
			Action: func(c *cli.Context) error {
				return cmdQuery(c)
			},
		},
		{
			Name:  "get",
			Usage: "Retrieve a [BF_TX] by its ID (Parameters: BF_TX id)",
			Action: func(c *cli.Context) error {
				return cmdGetBfTx(c)
			},
		},
		{
			Name:  "append",
			Usage: "Append a new BF_TX to an existing BF_TX (Parameters: JSON Filepath, BF_TX id)",
			Action: func(c *cli.Context) error {
				return cmdAppendBfTx(c)
			},
		},
		{
			Name:  "state",
			Usage: "Get the current state of a determined BF_TX (Parameters: BF_TX id)",
			Action: func(c *cli.Context) error {
				return cmdStateBfTx(c)
			},
		},
		{
			Name:  "total",
			Usage: "Query the total of BF_TX in DB (Parameters: none)",
			Action: func(c *cli.Context) error {
				return cmdTotalBfTx(c)
			},
		},
		{
			Name:  "echo",
			Usage: "Print clearly a BF_TX (Parameters: BF_TX id)",
			Action: func(c *cli.Context) error {
				return cmdPrintBfTx(c)
			},
		},
		{
			Name:  "exit",
			Usage: "Leaves the program. (Parameters: none)",
			Action: func(c *cli.Context) {
				os.Exit(0)
			},
		},
		{
			Name:  "saberenc",
			Usage: "prototype of saber encoding service.",
			Action: func(c *cli.Context) error {
				return cmdSaberEnc(c)
			},
		},
		{
			Name:  "saberdcp",
			Usage: "prototype of saber decoding service.",
			Action: func(c *cli.Context) error {
				return cmdSaberDcp(c)
			},
		},
	}
	app.Before = before
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err.Error())
	}

}

func getFunctionName(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}

// simpleLogger writes errors and the function name that generated the error to bftx.log
func simpleLogger(i interface{}, currentError error) {
	// If the file doesn't exist, create it, or append to the file
	f, err := os.OpenFile("bftx.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
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

// transLogger writes errors, the function name that generated the error, and the transaction body to bftx.log
func transLogger(i interface{}, currentError error, transactionBody bf_tx.BF_TX) {
	// If the file doesn't exist, create it, or append to the file
	f, err := os.OpenFile("bftx.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	tb, err := bf_tx.BFTXContent(transactionBody)
	if _, err := f.Write([]byte(time.Now().Format("2006/01/02 15:04") + ", " + getFunctionName(i) + ", " + currentError.Error() + ", " + tb + "\n\n")); err != nil {
		log.Fatal(err)
	}
	if err := f.Close(); err != nil {
		log.Fatal(err)
	}
}

func before(c *cli.Context) error {
	introduction(c)
	if client == nil {
		var err error
		client, err = abcicli.NewClient(c.GlobalString("address"), c.GlobalString("call"), true)
		client.Start()
		if err != nil {
			log.Fatal(err.Error())
			simpleLogger(before, err)
		}

		handlers.TendermintClient = client
	}

	return nil
}

// badCmd is called when we invoke with an invalid first argument (just for console for now)
func badCmd(c *cli.Context, cmd string) {
	fmt.Println("Unknown command:", cmd)
	fmt.Println("Please try one of the following:")
	fmt.Println("")
	cli.DefaultAppComplete(c)
}

// Generates new Args array based off of previous call args to maintain flag persistence
func persistentArgs(line []byte) []string {

	// generate the arguments to run from original os.Args
	// to maintain flag arguments
	args := os.Args
	args = args[:len(args)-1] // remove the previous command argument

	if len(line) > 0 { //prevents introduction of extra space leading to argument parse errors
		args = append(args, strings.Split(string(line), " ")...)
	}
	return args
}

func cmdGenerateBftxID(bftx bf_tx.BF_TX) (string, error) {
	// BlockID defines the unique ID of a block as its Hash and its PartSetHeader
	salt, err := getBlockAppHash()
	if err != nil {
		transLogger(cmdGenerateBftxID, err, bftx)
		return "", err
	}

	// Hash BF_TX Object
	hash, err := bf_tx.HashBFTX(bftx)
	if err != nil {
		transLogger(cmdGenerateBftxID, err, bftx)
		return "", err
	}

	// Generate BF_TX id
	bftxID := bf_tx.GenerateBFTXUID(hash, salt)

	return bftxID, nil
}

func getBlockAppHash() ([]byte, error) {
	resInfo, err := client.InfoSync(types.RequestInfo{})
	if err != nil {
		simpleLogger(getBlockAppHash, err)
		return nil, err
	}

	return resInfo.LastBlockAppHash, nil
}

func cmdBatch(app *cli.App, c *cli.Context) error {
	bufReader := bufio.NewReader(os.Stdin)
	for {
		line, more, err := bufReader.ReadLine()
		if more {
			return errors.New("Input line is too long")
		} else if err == io.EOF {
			break
		} else if len(line) == 0 {
			continue
		} else if err != nil {
			return err
		}

		args := persistentArgs(line)
		app.Run(args) //cli prints error within its func call
	}
	return nil
}

func cmdConsole(app *cli.App, c *cli.Context) error {
	// don't hard exit on mistyped commands (eg. check vs check_tx)
	app.CommandNotFound = badCmd

	for {
		fmt.Printf("\n> ")
		bufReader := bufio.NewReader(os.Stdin)
		line, more, err := bufReader.ReadLine()
		if more {
			return errors.New("Input is too long")
		} else if err != nil {
			return err
		}

		args := persistentArgs(line)
		app.Run(args) //cli prints error within its func call
	}
}

func cmdSaberEnc(c *cli.Context) error {
	args := c.Args()
	if len(args) != 1 {
		return errors.New("Command sign takes 1 argument")
	}
	// TODO: Change the arguments so it can specify the saber encoding parameters
	// CUrrent: it only takes the default encoding configuration

	// Get a BF_TX by id
	oldbftx, err := leveldb.GetBfTx(args[0])
	if err != nil {
		simpleLogger(cmdSaberEnc, err)
		return err
	}
	// In the long term, this conversion is unnecessary, and it makes the program to be less efficient.
	nwbftx, err := saberservice.BftxStructConverstionON(&oldbftx)
	if err != nil {
		log.Fatalf("Conversion error, can not convert old bftx to new bftx structure")
		simpleLogger(cmdSaberEnc, err)
		return err
	}
	st := saberservice.SaberDefaultInput()
	saberbftx, err := saberservice.SaberEncoding(nwbftx, st)
	if err != nil {
		simpleLogger(cmdSaberEnc, err)
		return err
	}
	bftxold, err := saberservice.BftxStructConverstionNO(saberbftx)
	//update the encoded transaction to database
	// Get the BF_TX content in string format
	content, err := bf_tx.BFTXContent(*bftxold)
	if err != nil {
		simpleLogger(cmdSaberEnc, err)
		return err
	}

	// Update on DB
	err = leveldb.RecordOnDB(string(bftxold.Id), content)
	if err != nil {
		simpleLogger(cmdSaberEnc, err)
		return err
	}
	// fmt.Printf("\nSaber encryption result: \n%+v\n", content)
	return nil
}

func cmdSaberDcp(c *cli.Context) error {
	args := c.Args()
	if len(args) != 1 {
		return errors.New("Command sign takes 1 argument")
	}
	// TODO: Change the arguments so it can specify the saber encoding parameters
	// CUrrent: it only takes the default encoding configuration

	// Get a BF_TX by id
	oldbftx, err := leveldb.GetBfTx(args[0])
	if err != nil {
		transLogger(cmdSaberDcp, err, oldbftx)
		return err
	}
	nwbftx, err := saberservice.BftxStructConverstionON(&oldbftx)
	if err != nil {
		log.Fatalf("Conversion error, can not convert old bftx to new bftx structure")
		transLogger(cmdSaberDcp, err, oldbftx)
		return err
	}
	st := saberservice.SaberDefaultInput()
	saberbftx, err := saberservice.SaberDecoding(nwbftx, st)
	if err != nil {
		transLogger(cmdSaberDcp, err, oldbftx)
		return err
	}
	bftxold, err := saberservice.BftxStructConverstionNO(saberbftx)
	//update the encoded transaction to database
	// Get the BF_TX content in string format
	content, err := bf_tx.BFTXContent(*bftxold)
	if err != nil {
		transLogger(cmdSaberDcp, err, oldbftx)
		return err
	}

	// Update on DB
	err = leveldb.RecordOnDB(string(bftxold.Id), content)
	if err != nil {
		transLogger(cmdSaberDcp, err, oldbftx)
		return err
	}
	// fmt.Printf("\nSaber decryption result: \n%+v\n", content)
	return nil
}

func cmdSaberEncTest(c *cli.Context) error {
	st := saberservice.Saberinputcli(nil)
	bftx, err := saberservice.SaberEncodingTestCase(st)
	if err != nil {
		simpleLogger(cmdSaberEncTest, err)
		return err
	}
	fmt.Print(bftx)
	return nil
}

func cmdSaberDcpTest(c *cli.Context) error {
	st := saberservice.Saberinputcli(nil)
	bfenc, err := saberservice.SaberEncodingTestCase(st)
	bftx, err := saberservice.SaberDecoding(bfenc, st)
	if err != nil {
		simpleLogger(cmdSaberDcpTest, err)
		return err
	}
	fmt.Print(bftx)
	return nil

}

// Get some info from the application
func cmdInfo(c *cli.Context) error {
	resInfo, err := client.InfoSync(types.RequestInfo{})
	if err != nil {
		simpleLogger(cmdInfo, err)
		return err
	}
	printResponse(c, response{
		Data:   resInfo.LastBlockAppHash,
		Result: resInfo.Data,
	})
	return nil
}

// Set an option on the application
func cmdSetOption(c *cli.Context) error {
	args := c.Args()
	if len(args) != 2 {
		return errors.New("Command set_option takes 2 arguments (key, value)")
	}
	resSetOption, err := client.SetOptionSync(types.RequestSetOption{
		Key:   args[0],
		Value: args[1],
	})
	if err != nil {
		simpleLogger(cmdSetOption, err)
		return err
	}

	printResponse(c, response{
		Log: resSetOption.Log,
	})
	return nil
}

// Verify the JSON imput against a BF_TX
func cmdVerifyBfTx(c *cli.Context) error {
	args := c.Args()
	if len(args) != 1 {
		return errors.New("Command verify takes 1 argument")
	}

	// Read JSON and instance the BF_TX structure
	jbftx, err := bf_tx.SetBFTX(c.GlobalString("json_path") + args[0])
	if err != nil {
		simpleLogger(cmdVerifyBfTx, err)
		return err
	}

	// Get the BF_TX old_content in string format
	jcontent, err := bf_tx.BFTXContent(jbftx)
	if err != nil {
		simpleLogger(cmdVerifyBfTx, err)
		return err
	}

	result, err := leveldb.Verify(jcontent)
	if err != nil {
		simpleLogger(cmdVerifyBfTx, err)
		return err
	}
	if result == nil {
		return errors.New("JSON content does not have a BF_TX associated.")
	}

	// Result
	printResponse(c, response{
		Result: "The BF_TX associated to JSON content is " + string(result),
	})

	return nil
}

// cmdValidateBfTx validates a BFTX
func cmdValidateBfTx(c *cli.Context) error {
	args := c.Args()
	if len(args) != 1 {
		return errors.New("Command validate takes 1 argument")
	}

	// Read JSON and instance the BF_TX structure
	bftx, err := bf_tx.SetBFTX(c.GlobalString("json_path") + args[0])
	if err != nil {
		transLogger(cmdValidateBfTx, err, bftx)
		return err
	}

	// Validate the BF_TX
	result, err := validator.ValidateBFTX(bftx)
	if err != nil {
		fmt.Println(result)
		transLogger(cmdValidateBfTx, err, bftx)
		return err
	}

	// Result
	printResponse(c, response{
		Result: result,
	})
	return nil
}

// Construct the Blockfreight™ Transaction [BF_TX]
func cmdConstructBfTx(c *cli.Context) error {
	args := c.Args()
	if len(args) != 1 {
		return errors.New("Command construct takes 1 argument")
	}

	// Read JSON and instance the BF_TX structure
	bftx, err := bf_tx.SetBFTX(c.GlobalString("json_path") + args[0])
	if err != nil {
		transLogger(cmdConstructBfTx, err, bftx)
		return err
	}

	newId, err := cmdGenerateBftxID(bftx)
	if err != nil {
		transLogger(cmdConstructBfTx, err, bftx)
		return err
	}

	bftx.Id = newId

	// Re-validate a BF_TX before create a BF_TX
	result, err := validator.ValidateBFTX(bftx)
	if err != nil {
		fmt.Println(result)
		transLogger(cmdConstructBfTx, err, bftx)
		return err
	}

	// Get the BF_TX content in string format
	content, err := bf_tx.BFTXContent(bftx)
	if err != nil {
		transLogger(cmdConstructBfTx, err, bftx)
		return err
	}

	// Save on DB
	err = leveldb.RecordOnDB(bftx.Id, content)
	if err != nil {
		transLogger(cmdConstructBfTx, err, bftx)
		return err
	}

	// Result
	printResponse(c, response{
		Result: "BF_TX Id: " + bftx.Id,
	})

	return nil
}

// Sign the Blockfreight™ Transaction [BF_TX]
func cmdSignBfTx(c *cli.Context) error {
	args := c.Args()
	if len(args) != 1 {
		return errors.New("Command sign takes 1 argument")
	}

	// Get a BF_TX by id
	bftx, err := leveldb.GetBfTx(args[0])
	if err != nil {
		transLogger(cmdSignBfTx, err, bftx)
		return err
	}
	if bftx.Verified {
		return errors.New("BF_TX already signed.")
	}

	// Sign BF_TX
	bftx, err = crypto.SignBFTX(bftx)
	if err != nil {
		transLogger(cmdSignBfTx, err, bftx)
		return err
	}

	// Get the BF_TX content in string format
	content, err := bf_tx.BFTXContent(bftx)
	if err != nil {
		transLogger(cmdSignBfTx, err, bftx)
		return err
	}

	// Update on DB
	err = leveldb.RecordOnDB(string(bftx.Id), content)
	if err != nil {
		transLogger(cmdSignBfTx, err, bftx)
		return err
	}

	// Result
	printResponse(c, response{
		Result: "BF_TX signed",
	})
	return nil
}

// Deliver a new BF_TX to application
func cmdBroadcastBfTx(c *cli.Context) error {
	args := c.Args()
	if len(args) != 1 {
		return errors.New("Command broadcast takes 1 argument")
	}

	// Get a BF_TX by id
	bftx, err := leveldb.GetBfTx(args[0])
	if err != nil {
		transLogger(cmdBroadcastBfTx, err, bftx)
		return err
	}
	if !bftx.Verified {
		return errors.New("BF_TX is not signed yet.")
	}
	if bftx.Transmitted {
		return errors.New("BF_TX already transmitted.")
	}

	// Change the boolean valud for Transmitted attribute
	bftx.Transmitted = true

	// Get the BF_TX content in string format
	content, err := bf_tx.BFTXContent(bftx)
	if err != nil {
		transLogger(cmdBroadcastBfTx, err, bftx)
		return err
	}

	// Update on DB
	err = leveldb.RecordOnDB(string(bftx.Id), content)
	if err != nil {
		transLogger(cmdBroadcastBfTx, err, bftx)
		return err
	}

	if err != nil {
		transLogger(cmdBroadcastBfTx, err, bftx)
		return err
	}

	rpcClient = rpc.NewHTTP(os.Getenv("LOCAL_RPC_CLIENT_ADDRESS"), "/websocket")
	err = rpcClient.Start()
	if err != nil {
		fmt.Println("Error when initializing rpcClient")
		log.Fatal(err.Error())

	}

	defer rpcClient.Stop()

	var tx tmTypes.Tx
	tx = []byte(content)

	resp, rpcErr := rpcClient.BroadcastTxSync(tx)
	if rpcErr != nil {
		fmt.Printf("%+v\n", rpcErr)
		transLogger(cmdBroadcastBfTx, rpcErr, bftx)
		return rpcErr
	}

	printResponse(c, response{
		Data: resp.Hash,
		Log:  resp.Log,
	})

	return nil
}

// Get application Merkle root hash
func cmdCommit(c *cli.Context) error {
	result, err := client.CommitSync()
	if err != nil {
		simpleLogger(cmdCommit, err)
		return err
	}

	printResponse(c, response{
		Data: result.Data,
		Log:  result.Log,
	})
	return nil
}

// Query application state
// TODO JCNM: Make request and response support all fields.
func cmdQuery(c *cli.Context) error {
	args := c.Args()
	if len(args) != 1 {
		return errors.New("Command query takes 1 argument")
	}

	rpcClient = rpc.NewHTTP(os.Getenv("LOCAL_RPC_CLIENT_ADDRESS"), "/websocket")
	err := rpcClient.Start()
	if err != nil {
		fmt.Println("Error when initializing rpcClient")
		log.Fatal(err.Error())
		simpleLogger(cmdQuery, err)
	}

	query := "bftx.id='" + args[0] + "'"

	resQuery, err := rpcClient.TxSearch(query, true)
	if err != nil {
		simpleLogger(cmdQuery, err)
		return err
	}

	defer rpcClient.Stop()

	if len(resQuery) > 0 {
		printResponse(c, response{
			Result: string(resQuery[0].Tx),
		})

		return nil
	}

	return errors.New("Blockfreight Transaction not found.")
}

// Return the output JSON
func cmdGetBfTx(c *cli.Context) error {
	args := c.Args()
	if len(args) != 1 {
		return errors.New("Command get takes 1 argument")
	}

	// Get a BF_TX by id
	bftx, err := leveldb.GetBfTx(args[0])
	if err != nil {
		transLogger(cmdGetBfTx, err, bftx)
		return err
	}

	// Get the BF_TX content in string format
	content, err := bf_tx.BFTXContent(bftx)
	if err != nil {
		transLogger(cmdGetBfTx, err, bftx)
		return err
	}

	// Result
	printResponse(c, response{
		Result: content,
	})
	return nil
}

// Append a new BF_TX to an existing BF_TX
func cmdAppendBfTx(c *cli.Context) error {
	args := c.Args()
	if len(args) != 2 {
		return errors.New("Command append takes 2 arguments")
	}

	// Get a BF_TX by id
	oldBftx, err := leveldb.GetBfTx(args[1])
	if err != nil {
		transLogger(cmdAppendBfTx, err, oldBftx)
		return err
	}

	// Query the total of BF_TX in DB
	// n, err := leveldb.Total()
	if err != nil {
		transLogger(cmdAppendBfTx, err, oldBftx)
		return err
	}

	// Read JSON and instance the BF_TX structure
	newBftx, err := bf_tx.SetBFTX(c.GlobalString("json_path") + args[0])
	if err != nil {
		return err

	}

	// Set the BF_TX id
	//newBftx.Id = n+1     //TODO JCNM: Solve concurrency problem

	// Update the BF_TX appended attribute of the old BF_TX
	oldBftx.Amendment = newBftx.Id

	// Get the BF_TX (old and new) content in string format
	newContent, err := bf_tx.BFTXContent(newBftx)
	if err != nil {
		transLogger(cmdAppendBfTx, err, newBftx)
		return err
	}
	oldContent, err := bf_tx.BFTXContent(oldBftx)
	if err != nil {
		transLogger(cmdAppendBfTx, err, newBftx)
		return err
	}

	// Save on DB
	err = leveldb.RecordOnDB(string(newBftx.Id), newContent)
	if err != nil {
		transLogger(cmdAppendBfTx, err, newBftx)
		return err
	}

	// Update on DB
	err = leveldb.RecordOnDB(string(oldBftx.Id), oldContent)
	if err != nil {
		transLogger(cmdAppendBfTx, err, newBftx)
		return err
	}

	//Result
	printResponse(c, response{
		Result: "BF_TX Id: " + string(newBftx.Id),
	})

	return nil
}

// Get the current state of a determined BF_TX
func cmdStateBfTx(c *cli.Context) error {
	args := c.Args()
	if len(args) != 1 {
		return errors.New("Command state takes 1 argument")
	}

	// Get a BF_TX by id
	bftx, err := leveldb.GetBfTx(args[0])
	if err != nil {
		transLogger(cmdStateBfTx, err, bftx)
		return err
	}

	// Result
	printResponse(c, response{
		Result: "BF_TX state: " + bf_tx.State(bftx),
	})
	return nil
}

func cmdPrintBfTx(c *cli.Context) error {
	args := c.Args()
	if len(args) != 1 {
		return errors.New("Command print takes 1 argument")
	}

	// Get a BF_TX by id
	bftx, err := leveldb.GetBfTx(args[0])
	if err != nil {
		transLogger(cmdPrintBfTx, err, bftx)
		return err
	}

	// Print the BF_TX clearly
	bf_tx.PrintBFTX(bftx)
	return nil
}

func cmdTotalBfTx(c *cli.Context) error {
	// Query the total of BF_TX in DB
	total, err := leveldb.Total()
	if err != nil {
		simpleLogger(cmdTotalBfTx, err)
		return err
	}

	// Result
	printResponse(c, response{
		Result: "Total BF_TX on BD: " + strconv.Itoa(total),
	})
	return nil
}

//--------------------------------------------------------------------------------

func printResponse(c *cli.Context, rsp response) {

	verbose := c.GlobalBool("verbose")

	if verbose {
		fmt.Println(">", c.Command.Name, strings.Join(c.Args(), " "))
	}

	if rsp.Result != "" {
		fmt.Printf("-> blockfreight result: %s\n", rsp.Result)
	}
	if len(rsp.Data) != 0 {
		//fmt.Printf("-> blockfreight data: %s\n", rsp.Data)
		fmt.Printf("-> data.hex: %X\n", rsp.Data)
	}
	if rsp.Log != "" {
		fmt.Printf("-> log: %s\n", rsp.Log)
	}

	if rsp.Query != nil {
		fmt.Printf("-> height: %d\n", rsp.Query.Height)
		if rsp.Query.Key != nil {
			fmt.Printf("-> key: %s\n", rsp.Query.Key)
			fmt.Printf("-> key.hex: %X\n", rsp.Query.Key)
		}
		if rsp.Query.Value != nil {
			fmt.Printf("-> value: %s\n", rsp.Query.Value)
			fmt.Printf("-> value.hex: %X\n", rsp.Query.Value)
		}
		if rsp.Query.Proof != nil {
			fmt.Printf("-> proof: %X\n", rsp.Query.Proof)
		}
	}

	if verbose {
		fmt.Println("")
	}

}

// NOTE: s is interpreted as a string unless prefixed with 0x
func stringOrHexToBytes(s string) ([]byte, error) {
	if len(s) > 2 && strings.ToLower(s[:2]) == "0x" {
		b, err := hex.DecodeString(s[2:])
		if err != nil {
			err = fmt.Errorf("Error decoding hex argument: %s", err.Error())
			simpleLogger(stringOrHexToBytes, err)
			return nil, err
		}
		return b, nil
	}

	if !strings.HasPrefix(s, "\"") || !strings.HasSuffix(s, "\"") {
		err := fmt.Errorf("Invalid string arg: \"%s\". Must be quoted or a \"0x\"-prefixed hex string", s)
		return nil, err
	}

	return []byte(s[1 : len(s)-1]), nil
}

func introduction(c *cli.Context) {
	fmt.Println("\n...........................................")
	fmt.Println("Blockfreight™ Go App")
	fmt.Println("Address " + c.GlobalString("address"))
	fmt.Println("API Address http://localhost:12345")
	fmt.Println("BFT Implementation:  " + c.GlobalString("call"))
	fmt.Println("...........................................\n")
	/*name := "Blockfreight Community"
	  if c.NArg() > 0 {
	    name = c.Args().Get(0)
	  }
	  if c.String("lang") == "ES" { //ISO 639-1
	    fmt.Println("Hola", name)
	  } else {
	    fmt.Println("Hello", name)
	  }*/
}

// =================================================
// Blockfreight™ | The blockchain of global freight.
// =================================================

// BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB
// BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB
// BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB
// BBBBBBB                    BBBBBBBBBBBBBBBBBBB
// BBBBBBB                       BBBBBBBBBBBBBBBB
// BBBBBBB                        BBBBBBBBBBBBBBB
// BBBBBBB       BBBBBBBBB        BBBBBBBBBBBBBBB
// BBBBBBB       BBBBBBBBB        BBBBBBBBBBBBBBB
// BBBBBBB       BBBBBBB         BBBBBBBBBBBBBBBB
// BBBBBBB                     BBBBBBBBBBBBBBBBBB
// BBBBBBB                        BBBBBBBBBBBBBBB
// BBBBBBB       BBBBBBBBBB        BBBBBBBBBBBBBB
// BBBBBBB       BBBBBBBBBBB       BBBBBBBBBBBBBB
// BBBBBBB       BBBBBBBBBB        BBBBBBBBBBBBBB
// BBBBBBB       BBBBBBBBB        BBB       BBBBB
// BBBBBBB                       BBBB       BBBBB
// BBBBBBB                    BBBBBBB       BBBBB
// BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB
// BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB
// BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB

// ==================================================
// Blockfreight™ | The blockchain for global freight.
// ==================================================
