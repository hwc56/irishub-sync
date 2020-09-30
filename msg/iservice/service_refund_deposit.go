package iservice

import (
	. "github.com/irisnet/irishub-sync/util/constant"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/irisnet/irishub-sync/store/document"
	"github.com/irisnet/irishub-sync/types"
	"github.com/irisnet/irishub-sync/store"
)

type (
	DocMsgRefundServiceDeposit struct {
		ServiceName string `bson:"service_name" yaml:"service_name"`
		Provider    string `bson:"provider" yaml:"provider"`
		Owner       string `bson:"owner" yaml:"owner"`
	}
)

func (m *DocMsgRefundServiceDeposit) Type() string {
	return TxTypeRefundServiceDeposit
}

func (m *DocMsgRefundServiceDeposit) BuildMsg(v interface{}) {
	msg := v.(*types.MsgRefundServiceDeposit)

	m.ServiceName = msg.ServiceName
	m.Provider = msg.Provider.String()
	m.Owner = msg.Owner.String()
}

func (m *DocMsgRefundServiceDeposit) HandleTxMsg(msgData sdk.Msg, tx *document.CommonTx) *document.CommonTx {

	m.BuildMsg(msgData)
	tx.Msgs = append(tx.Msgs, document.DocTxMsg{
		Type: m.Type(),
		Msg:  m,
	})
	tx.Addrs = append(tx.Addrs, m.Provider, m.Owner)
	tx.Types = append(tx.Types, m.Type())
	if len(tx.Msgs) > 1 {
		return tx
	}
	tx.Type = m.Type()
	if len(tx.Signers) > 0 {
		tx.From = tx.Signers[0].AddrBech32
	}
	tx.To = ""
	tx.Amount = []store.Coin{}
	return tx
}
