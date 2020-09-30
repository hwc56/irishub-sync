package iservice

import (
	"encoding/hex"
	. "github.com/irisnet/irishub-sync/util/constant"
	"github.com/irisnet/irishub-sync/store/document"
	"github.com/irisnet/irishub-sync/types"
	"github.com/irisnet/irishub-sync/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type (
	DocMsgUpdateRequestContext struct {
		RequestContextId  string   `bson:"request_context_id" yaml:"request_context_id"`
		Providers         []string `bson:"providers" yaml:"providers"`
		Consumer          string   `bson:"consumer" yaml:"consumer"`
		ServiceFeeCap     Coins    `bson:"service_fee_cap" yaml:"service_fee_cap"`
		Timeout           int64    `bson:"timeout" yaml:"timeout"`
		RepeatedFrequency uint64   `bson:"repeated_frequency" yaml:"repeated_frequency"`
		RepeatedTotal     int64    `bson:"repeated_total" yaml:"repeated_total"`
	}
)

func (m *DocMsgUpdateRequestContext) Type() string {
	return TxTypeUpdateRequestContext
}

func (m *DocMsgUpdateRequestContext) BuildMsg(v interface{}) {
	msg := v.(*types.MsgUpdateRequestContext)

	loadProviders := func() (ret []string) {
		for _, one := range msg.Providers {
			ret = append(ret, one.String())
		}
		return
	}

	var coins Coins
	for _, one := range msg.ServiceFeeCap {
		coins = append(coins, Coin{Denom: one.Denom, Amount: one.Amount.String()})
	}

	m.RequestContextId = hex.EncodeToString(msg.RequestContextId)
	m.Providers = loadProviders()
	m.Consumer = msg.Consumer.String()
	m.ServiceFeeCap = coins
	m.Timeout = msg.Timeout
	m.RepeatedFrequency = msg.RepeatedFrequency
	m.RepeatedTotal = msg.RepeatedTotal
}

func (m *DocMsgUpdateRequestContext) HandleTxMsg(msgData sdk.Msg, tx *document.CommonTx) *document.CommonTx {

	m.BuildMsg(msgData)
	tx.Msgs = append(tx.Msgs, document.DocTxMsg{
		Type: m.Type(),
		Msg:  m,
	})
	tx.Addrs = append(tx.Addrs, m.Consumer)
	tx.Addrs = append(tx.Addrs, m.Providers...)
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
