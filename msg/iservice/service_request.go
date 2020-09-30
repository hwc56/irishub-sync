package iservice

import (
	. "github.com/irisnet/irishub-sync/util/constant"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/irisnet/irishub-sync/store/document"
	"github.com/irisnet/irishub-sync/store"
	"github.com/irisnet/irishub-sync/types"
)

type (
	DocMsgCallService struct {
		ServiceName       string   `bson:"service_name"`
		Providers         []string `bson:"providers"`
		Consumer          string   `bson:"consumer"`
		Input             string   `bson:"input"`
		ServiceFeeCap     Coins    `bson:"service_fee_cap"`
		Timeout           int64    `bson:"timeout"`
		SuperMode         bool     `bson:"super_mode"`
		Repeated          bool     `bson:"repeated"`
		RepeatedFrequency uint64   `bson:"repeated_frequency"`
		RepeatedTotal     int64    `bson:"repeated_total"`
	}
)

func (m *DocMsgCallService) Type() string {
	return TxTypeCallService
}

func (m *DocMsgCallService) BuildMsg(msg interface{}) {
	v := msg.(*types.MsgCallService)

	loadProviders := func() (ret []string) {
		for _, one := range v.Providers {
			ret = append(ret, one.String())
		}
		return
	}

	var coins Coins
	for _, one := range v.ServiceFeeCap {
		coins = append(coins, Coin{Denom: one.Denom, Amount: one.Amount.String()})
	}
	m.ServiceName = v.ServiceName
	m.Providers = loadProviders()
	m.Consumer = v.Consumer.String()
	m.Input = v.Input
	m.ServiceFeeCap = coins
	m.Timeout = v.Timeout
	//m.Input = hex.EncodeToString(v.Input)
	m.SuperMode = v.SuperMode
	m.Repeated = v.Repeated
	m.RepeatedFrequency = v.RepeatedFrequency
	m.RepeatedTotal = v.RepeatedTotal
}

func (m *DocMsgCallService) HandleTxMsg(msgData sdk.Msg, tx *document.CommonTx) *document.CommonTx {

	m.BuildMsg(msgData)
	tx.Msgs = append(tx.Msgs, document.DocTxMsg{
		Type: m.Type(),
		Msg:  m,
	})
	tx.Addrs = append(tx.Addrs, m.Providers...)
	tx.Addrs = append(tx.Addrs, m.Consumer)
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
	tx.Addrs = append(tx.Addrs, m.Providers...)
	tx.Addrs = append(tx.Addrs, m.Consumer)
	return tx
}
