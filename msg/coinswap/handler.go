package coinswap

import (
	"github.com/irisnet/irishub-sync/store/document"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/irisnet/irishub-sync/types"
	"github.com/irisnet/irishub-sync/util/constant"
	"github.com/irisnet/irishub-sync/store"
)

func HandleTxMsg(msgData sdk.Msg, docTx *document.CommonTx) (*document.CommonTx, bool) {
	ok := true
	switch msgData.Type() {
	case new(types.MsgAddLiquidity).Type():
		txMsg := DocTxMsgAddLiquidity{}
		txMsg.BuildMsg(msgData)
		docTx.Msgs = append(docTx.Msgs, document.DocTxMsg{
			Type: txMsg.Type(),
			Msg:  &txMsg,
		})
		docTx.Addrs = append(docTx.Addrs, txMsg.Sender)
		docTx.Types = append(docTx.Types, txMsg.Type())
		if len(docTx.Msgs) > 1 {
			return docTx, true
		}
		docTx.From = txMsg.Sender
		docTx.To = ""
		docTx.Amount = store.Coins{txMsg.MaxToken}
		docTx.Type = constant.TxTypeAddLiquidity
	case new(types.MsgRemoveLiquidity).Type():

		txMsg := DocTxMsgRemoveLiquidity{}
		txMsg.BuildMsg(msgData)
		docTx.Msgs = append(docTx.Msgs, document.DocTxMsg{
			Type: txMsg.Type(),
			Msg:  &txMsg,
		})
		docTx.Addrs = append(docTx.Addrs, txMsg.Sender)
		docTx.Types = append(docTx.Types, txMsg.Type())
		if len(docTx.Msgs) > 1 {
			return docTx, true
		}
		docTx.From = txMsg.Sender
		docTx.To = ""
		docTx.Amount = store.Coins{txMsg.WithdrawLiquidity}
		docTx.Type = constant.TxTypeRemoveLiquidity
	case new(types.MsgSwapOrder).Type():

		txMsg := DocTxMsgSwapOrder{}
		txMsg.BuildMsg(msgData)
		docTx.Msgs = append(docTx.Msgs, document.DocTxMsg{
			Type: txMsg.Type(),
			Msg:  &txMsg,
		})
		docTx.Addrs = append(docTx.Addrs, txMsg.Input.Address, txMsg.Output.Address)
		docTx.Types = append(docTx.Types, txMsg.Type())
		if len(docTx.Msgs) > 1 {
			return docTx, true
		}
		docTx.From = txMsg.Input.Address
		docTx.To = txMsg.Output.Address
		docTx.Amount = store.Coins{txMsg.Input.Coin}
		docTx.Type = constant.TxTypeSwapOrder
	default:
		ok = false
	}
	return docTx, ok
}
