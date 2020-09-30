package htlc

import (
	"github.com/irisnet/irishub-sync/store/document"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/irisnet/irishub-sync/types"
	"github.com/irisnet/irishub-sync/util/constant"
)

func HandleTxMsg(msgData sdk.Msg, docTx *document.CommonTx) (*document.CommonTx, bool) {
	ok := true
	switch msgData.Type() {
	case new(types.MsgCreateHTLC).Type():
		msg := msgData.(*types.MsgCreateHTLC)

		txMsg := DocTxMsgCreateHTLC{}
		txMsg.BuildMsg(msg)
		docTx.Msgs = append(docTx.Msgs, document.DocTxMsg{
			Type: txMsg.Type(),
			Msg:  &txMsg,
		})
		docTx.Addrs = append(docTx.Addrs, txMsg.Sender, txMsg.To)
		docTx.Types = append(docTx.Types, txMsg.Type())
		if len(docTx.Msgs) > 1 {
			return docTx, true
		}
		docTx.From = msg.Sender.String()
		docTx.To = msg.To.String()
		docTx.Amount = types.ParseCoins(msg.Amount.String())
		docTx.Type = constant.TxTypeCreateHTLC
	case new(types.MsgClaimHTLC).Type():
		msg := msgData.(*types.MsgClaimHTLC)

		txMsg := DocTxMsgClaimHTLC{}
		txMsg.BuildMsg(msg)
		docTx.Msgs = append(docTx.Msgs, document.DocTxMsg{
			Type: txMsg.Type(),
			Msg:  &txMsg,
		})
		docTx.Addrs = append(docTx.Addrs, txMsg.Sender)
		docTx.Types = append(docTx.Types, txMsg.Type())
		if len(docTx.Msgs) > 1 {
			return docTx, true
		}
		docTx.From = msg.Sender.String()
		docTx.To = ""
		docTx.Type = constant.TxTypeClaimHTLC
	case new(types.MsgRefundHTLC).Type():
		msg := msgData.(*types.MsgRefundHTLC)

		txMsg := DocTxMsgRefundHTLC{}
		txMsg.BuildMsg(msg)
		docTx.Msgs = append(docTx.Msgs, document.DocTxMsg{
			Type: txMsg.Type(),
			Msg:  &txMsg,
		})
		docTx.Addrs = append(docTx.Addrs, txMsg.Sender)
		docTx.Types = append(docTx.Types, txMsg.Type())
		if len(docTx.Msgs) > 1 {
			return docTx, true
		}
		docTx.From = msg.Sender.String()
		docTx.To = ""
		docTx.Type = constant.TxTypeRefundHTLC
	default:
		ok = false
	}
	return docTx, ok
}
