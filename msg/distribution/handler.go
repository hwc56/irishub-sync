package distribution

import (
	"github.com/irisnet/irishub-sync/store/document"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/irisnet/irishub-sync/types"
	"github.com/irisnet/irishub-sync/util/constant"
)

func HandleTxMsg(msgData sdk.Msg, docTx *document.CommonTx) (*document.CommonTx, bool) {
	ok := true
	switch msgData.Type() {
	case new(types.MsgSetWithdrawAddress).Type():

		txMsg := DocTxMsgSetWithdrawAddress{}
		txMsg.BuildMsg(msgData)
		docTx.Msgs = append(docTx.Msgs, document.DocTxMsg{
			Type: txMsg.Type(),
			Msg:  &txMsg,
		})
		docTx.Addrs = append(docTx.Addrs, txMsg.DelegatorAddr, txMsg.WithdrawAddr)
		docTx.Types = append(docTx.Types, txMsg.Type())
		if len(docTx.Msgs) > 1 {
			return docTx, true
		}
		docTx.From = txMsg.DelegatorAddr
		docTx.To = txMsg.WithdrawAddr
		docTx.Type = constant.TxTypeSetWithdrawAddress
	case new(types.MsgWithdrawDelegatorReward).Type():

		txMsg := DocTxMsgWithdrawDelegatorReward{}
		txMsg.BuildMsg(msgData)
		docTx.Msgs = append(docTx.Msgs, document.DocTxMsg{
			Type: txMsg.Type(),
			Msg:  &txMsg,
		})
		docTx.Addrs = append(docTx.Addrs, txMsg.DelegatorAddr, txMsg.ValidatorAddr)
		docTx.Types = append(docTx.Types, txMsg.Type())
		if len(docTx.Msgs) > 1 {
			return docTx, true
		}
		docTx.From = txMsg.DelegatorAddr
		docTx.To = txMsg.ValidatorAddr
		docTx.Type = constant.TxTypeWithdrawDelegatorReward

	case new(types.MsgFundCommunityPool).Type():

		txMsg := DocTxMsgFundCommunityPool{}
		txMsg.BuildMsg(msgData)
		docTx.Msgs = append(docTx.Msgs, document.DocTxMsg{
			Type: txMsg.Type(),
			Msg:  &txMsg,
		})
		docTx.Addrs = append(docTx.Addrs, txMsg.Depositor)
		docTx.Types = append(docTx.Types, txMsg.Type())
		if len(docTx.Msgs) > 1 {
			return docTx, true
		}
		docTx.From = txMsg.Depositor
		docTx.Amount = txMsg.Amount
		docTx.Type = constant.TxTypeMsgFundCommunityPool
	case new(types.MsgWithdrawValidatorCommission).Type():

		txMsg := DocTxMsgWithdrawValidatorCommission{}
		txMsg.BuildMsg(msgData)
		docTx.Msgs = append(docTx.Msgs, document.DocTxMsg{
			Type: txMsg.Type(),
			Msg:  &txMsg,
		})
		docTx.Addrs = append(docTx.Addrs, txMsg.ValidatorAddr)
		docTx.Types = append(docTx.Types, txMsg.Type())
		if len(docTx.Msgs) > 1 {
			return docTx, true
		}
		docTx.From = txMsg.ValidatorAddr
		docTx.Type = constant.TxTypeMsgWithdrawValidatorCommission
	default:
		ok = false
	}
	return docTx, ok
}
