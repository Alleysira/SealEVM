package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"github.com/SealSC/SealEVM"
	"github.com/SealSC/SealEVM/crypto/hashes"
	"github.com/SealSC/SealEVM/environment"
	"github.com/SealSC/SealEVM/evmInt256"
	"github.com/SealSC/SealEVM/instructions"
	"math/big"
	"strings"
	"time"
)

// store result to memStorage
func storeResult(result *SealEVM.ExecuteResult, storage *memStorage) {
	for addr, cache := range result.StorageCache.CachedData {
		for key, v := range cache {
			storage.storage[addr+key] = v.Bytes()
		}
	}
}

// create a new evm
func newEvm(code []byte, callData []byte, caller []byte, ms *memStorage) *SealEVM.EVM {
	hash := hashes.Keccak256(code)
	hashInt := evmInt256.New(0)
	hashInt.SetBytes(hash)

	//same contract code has same address in this example
	cNamespace := evmInt256.New(0x0000000000000000000000007265636569766572)
	contract := environment.Contract{
		Namespace: cNamespace,
		Code:      code,
		Hash:      hashInt,
	}

	var callHash [32]byte
	copy(callHash[12:], caller)
	callerInt, _ := evmInt256.HashBytesToEVMInt(callHash)
	sender := new(big.Int)
	sender.SetString("1c7cd2d37ffd63856a5bd56a9af1643f2bcf545f", 16)
	evm := SealEVM.New(SealEVM.EVMParam{
		MaxStackDepth:  1024,
		ExternalStore:  ms,
		ResultCallback: nil,
		Context: &environment.Context{
			Block: environment.Block{
				ChainID:    evmInt256.New(9599),
				Coinbase:   evmInt256.New(0xabcd),
				Timestamp:  evmInt256.New(int64(time.Now().Second())),
				Number:     evmInt256.New(0),
				Difficulty: evmInt256.New(0),
				GasLimit:   evmInt256.New(0xffffff), // real one
				Hash:       evmInt256.New(0),
			},
			Contract: contract,
			Transaction: environment.Transaction{
				Origin:   evmInt256.FromBigInt(sender),
				GasPrice: evmInt256.New(0),
				GasLimit: evmInt256.New(0xffffff),
			},
			Message: environment.Message{
				Caller: callerInt,
				Value:  evmInt256.New(0),
				Data:   callData,
			},
		},
		GasSetting: instructions.DefaultGasSetting(),
	})

	return evm
}

func main() {
	//load SealEVM module
	p_deployCode := flag.String("code", "", "bytecode")
	p_calldata := flag.String("sig", "", "signature + calldata")
	flag.Parse()
	SealEVM.Load()

	//create memStorage
	ms := &memStorage{}
	ms.storage = make(map[string][]byte)
	ms.contracts = make(map[string][]byte)

	deployCode, codeerr := hex.DecodeString(*p_deployCode)
	if codeerr != nil {
		fmt.Println("bad bincode")
	}

	calldata, sigerr := hex.DecodeString(strings.TrimPrefix(*p_calldata, "0x"))
	if sigerr != nil {
		fmt.Println(*p_calldata, sigerr, "bad sigdata")
	}

	evm := newEvm(deployCode, calldata, caller, ms)
	ret, _ := evm.ExecuteContract(true)

	//store the result to ms
	storeResult(&ret, ms)

}
