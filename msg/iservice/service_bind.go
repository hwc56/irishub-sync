package iservice

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/irisnet/irishub-sync/types"
	. "github.com/irisnet/irishub-sync/util/constant"
	"github.com/irisnet/irishub-sync/store/document"
)

type (
	DocMsgBindService struct {
		ServiceName string `bson:"service_name"`
		Provider    string `bson:"provider"`
		Deposit     Coins  `bson:"deposit"`
		Pricing     string `bson:"pricing"`
		QoS         uint64 `bson:"qos"`
		Owner       string `bson:"owner"`
	}
)

func (m *DocMsgBindService) Type() string {
	return TxTypeBindService
}

func (m *DocMsgBindService) BuildMsg(v interface{}) {
	msg := v.(*types.MsgBindService)

	var coins Coins
	for _, one := range msg.Deposit {
		coins = append(coins, Coin{Denom: one.Denom, Amount: one.Amount.String()})
	}
	m.ServiceName = msg.ServiceName
	m.Provider = msg.Provider.String()
	m.Deposit = coins
	m.Pricing = msg.Pricing
	m.QoS = msg.QoS
	m.Owner = msg.Owner.String()
}

func (m *DocMsgBindService) HandleTxMsg(msgData sdk.Msg, tx *document.CommonTx) *document.CommonTx {

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
	tx.Amount = m.Deposit.Convert()
	return tx
}
