package nft

import (
	. "github.com/irisnet/irishub-sync/util/constant"
	"strings"
	"github.com/irisnet/irishub-sync/store/document"
	"github.com/irisnet/irishub-sync/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/irisnet/irishub-sync/store"
)

type (
	DocMsgNFTBurn struct {
		Sender string `bson:"sender"`
		Id     string `bson:"id"`
		Denom  string `bson:"denom"`
	}
)

func (m *DocMsgNFTBurn) Type() string {
	return TxTypeNFTBurn
}

func (m *DocMsgNFTBurn) BuildMsg(v interface{}) {
	msg := v.(*types.MsgBurnNFT)

	m.Sender = msg.Sender.String()
	m.Id = strings.ToLower(msg.Id)
	m.Denom = strings.ToLower(msg.Denom)
}

func (m *DocMsgNFTBurn) HandleTxMsg(msgData sdk.Msg, tx *document.CommonTx) *document.CommonTx {

	m.BuildMsg(msgData)
	tx.Msgs = append(tx.Msgs, document.DocTxMsg{
		Type: m.Type(),
		Msg:  m,
	})
	tx.Addrs = append(tx.Addrs, m.Sender)
	tx.Types = append(tx.Types, m.Type())
	if len(tx.Msgs) > 1 {
		return tx
	}
	tx.Type = m.Type()
	tx.From = m.Sender
	tx.To = ""
	tx.Amount = []store.Coin{}
	return tx
}
