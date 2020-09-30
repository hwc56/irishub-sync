// package for parse tx struct from binary data

package helper

import (
	"encoding/hex"
	"github.com/irisnet/irishub-sync/logger"
	"github.com/irisnet/irishub-sync/store"
	"github.com/irisnet/irishub-sync/store/document"
	"github.com/irisnet/irishub-sync/types"
	"strings"
	"time"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	"github.com/cosmos/cosmos-sdk/x/auth/signing"
	"github.com/irisnet/irishub-sync/cdc"
)

func ParseTx(txBytes types.Tx, block *types.Block) *document.CommonTx {
	var (
		methodName = "ParseTx"
		docTx      *document.CommonTx
		gasPrice   float64
		actualFee  store.ActualFee
		signers    []document.Signer
		signerAddr string
		addrs      []string
	)

	Tx, err := cdc.GetTxDecoder()(txBytes)
	if err != nil {
		logger.Error(err.Error())
		return docTx
	}

	height := block.Height
	blockTime := block.Time
	txHash := BuildHex(txBytes.Hash())

	authTx := Tx.(signing.Tx)
	fee := types.BuildFee(authTx.GetFee(), authTx.GetGas())
	memo := authTx.GetMemo()

	// get tx signers
	if len(authTx.GetSigners()) > 0 {
		for _, signature := range authTx.GetSigners() {
			signer := document.Signer{}
			signer.AddrHex = hex.EncodeToString([]byte(signature.String()))
			if addrBech32, err := bech32.ConvertAndEncode(types.Bech32AccountAddrPrefix, signature.Bytes()); err != nil {
				logger.Error("convert account addr from hex to bech32 fail",
					logger.String("addrHex", signature.String()), logger.String("err", err.Error()))
			} else {
				signer.AddrBech32 = addrBech32
				signerAddr = addrBech32
			}
			addrs = append(addrs, signer.AddrBech32)
			signers = append(signers, signer)
		}
	}

	// get tx status, gasUsed, gasPrice and actualFee from tx result
	status, result, err := QueryTxResult(txBytes.Hash())
	if err != nil {
		logger.Error("get txResult err", logger.String("method", methodName), logger.String("err", err.Error()))
	}
	log := result.Log
	gasUsed := Min(result.GasUsed, fee.Gas)
	if len(fee.Amount) > 0 {
		gasPrice = fee.Amount[0].Amount / float64(fee.Gas)
		actualFee = store.ActualFee{
			Denom:  fee.Amount[0].Denom,
			Amount: float64(gasUsed) * gasPrice,
		}
	} else {
		gasPrice = 0
		actualFee = store.ActualFee{}
	}

	msgs := authTx.GetMsgs()
	if len(msgs) <= 0 {
		logger.Error("can't get msgs", logger.String("method", methodName))
		return docTx
	}

	docTx = &document.CommonTx{
		Height:    height,
		Time:      blockTime,
		TxHash:    txHash,
		Fee:       fee,
		Memo:      memo,
		Status:    status,
		Code:      result.Code,
		Log:       log,
		GasUsed:   gasUsed,
		GasWanted: result.GasUsed,
		GasPrice:  gasPrice,
		ActualFee: actualFee,
		Events:    parseEvents(result),
		Signers:   signers,
		Signer:    signerAddr,
		TimeUnix:  blockTime.Unix(),
		Addrs:     addrs,
	}
	for _, msgData := range msgs {
		if len(msgData.GetSigners()) == 0 {
			continue
		}
		if docInfo, ok := HandleMsg(msgData, docTx); ok {
			docTx = docInfo
		}
	}

	docTx.Addrs = removeDuplicatesFromSlice(docTx.Addrs)
	docTx.Types = removeDuplicatesFromSlice(docTx.Types)

	return docTx
}

func removeDuplicatesFromSlice(data []string) (result []string) {
	tempAddrsSet := make(map[string]string, len(data))
	for _, val := range data {
		if _, ok := tempAddrsSet[val]; ok || val == "" {
			continue
		}
		tempAddrsSet[val] = val
	}
	for one := range tempAddrsSet {
		result = append(result, one)
	}
	return
}

func parseEvents(result types.ResponseDeliverTx) []document.Event {

	var events []document.Event
	for _, val := range result.GetEvents() {
		one := document.Event{
			Type: val.Type,
		}
		for _, attr := range val.Attributes {
			one.Attributes = append(one.Attributes, document.Attribute{Key: string(attr.Key), Value: string(attr.Value)})
		}
		events = append(events, one)
	}

	return events
}

func BuildHex(bytes []byte) string {
	return strings.ToUpper(hex.EncodeToString(bytes))
}

// get tx status and log by query txHash
func QueryTxResult(txHash []byte) (string, types.ResponseDeliverTx, error) {
	var resDeliverTx types.ResponseDeliverTx
	status := document.TxStatusSuccess

	client := GetClient()
	defer client.Release()

	res, err := client.Tx(txHash, false)
	if err != nil {
		// try again
		time.Sleep(time.Duration(1) * time.Second)
		if res, err := client.Tx(txHash, false); err != nil {
			return "unknown", resDeliverTx, err
		} else {
			resDeliverTx = res.TxResult
		}
	} else {
		resDeliverTx = res.TxResult
	}

	if resDeliverTx.Code != 0 {
		status = document.TxStatusFail
	}

	return status, resDeliverTx, nil
}
