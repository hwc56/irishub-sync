package guardian

import (
	"github.com/irisnet/irishub-sync/store/document"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/irisnet/irishub-sync/types"
	"github.com/irisnet/irishub-sync/util/constant"
)

func HandleTxMsg(msgData sdk.Msg, docTx *document.CommonTx) (*document.CommonTx, bool) {
	ok := true
	switch msgData.Type() {
	case new(types.MsgAddProfiler).Type():

		txMsg := DocTxMsgAddProfiler{}
		txMsg.BuildMsg(msgData)
		docTx.Msgs = append(docTx.Msgs, document.DocTxMsg{
			Type: txMsg.Type(),
			Msg:  &txMsg,
		})
		docTx.Addrs = append(docTx.Addrs, txMsg.Address, txMsg.AddedBy)
		docTx.Types = append(docTx.Types, txMsg.Type())
		if len(docTx.Msgs) > 1 {
			return docTx, true
		}
		docTx.From = txMsg.AddGuardian.AddedBy
		docTx.To = txMsg.AddGuardian.Address
		docTx.Type = constant.TxTypeAddProfiler

	case new(types.MsgAddTrustee).Type():
		txMsg := DocTxMsgAddTrustee{}
		txMsg.BuildMsg(msgData)
		docTx.Msgs = append(docTx.Msgs, document.DocTxMsg{
			Type: txMsg.Type(),
			Msg:  &txMsg,
		})
		docTx.Addrs = append(docTx.Addrs, txMsg.Address, txMsg.AddedBy)
		docTx.Types = append(docTx.Types, txMsg.Type())
		if len(docTx.Msgs) > 1 {
			return docTx, true
		}
		docTx.From = txMsg.AddGuardian.AddedBy
		docTx.To = txMsg.AddGuardian.Address
		docTx.Type = constant.TxTypeAddTrustee

	case new(types.MsgDeleteTrustee).Type():
		txMsg := DocTxMsgDeleteTrustee{}
		txMsg.BuildMsg(msgData)
		docTx.Msgs = append(docTx.Msgs, document.DocTxMsg{
			Type: txMsg.Type(),
			Msg:  &txMsg,
		})
		docTx.Addrs = append(docTx.Addrs, txMsg.DeletedBy, txMsg.Address)
		docTx.Types = append(docTx.Types, txMsg.Type())
		if len(docTx.Msgs) > 1 {
			return docTx, true
		}
		docTx.From = txMsg.DeleteGuardian.DeletedBy
		docTx.To = txMsg.DeleteGuardian.Address
		docTx.Type = constant.TxTypeDeleteTrustee

	case new(types.MsgDeleteProfiler).Type():
		txMsg := DocTxMsgDeleteProfiler{}
		txMsg.BuildMsg(msgData)
		docTx.Msgs = append(docTx.Msgs, document.DocTxMsg{
			Type: txMsg.Type(),
			Msg:  &txMsg,
		})
		docTx.Addrs = append(docTx.Addrs, txMsg.DeletedBy, txMsg.Address)
		docTx.Types = append(docTx.Types, txMsg.Type())
		if len(docTx.Msgs) > 1 {
			return docTx, true
		}
		docTx.From = txMsg.DeleteGuardian.DeletedBy
		docTx.To = txMsg.DeleteGuardian.Address
		docTx.Type = constant.TxTypeDeleteProfiler
	default:
		ok = false
	}
	return docTx, ok
}

