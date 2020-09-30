package gov

import (
	"github.com/irisnet/irishub-sync/store/document"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/irisnet/irishub-sync/types"
	"github.com/irisnet/irishub-sync/util/constant"
	"github.com/irisnet/irishub-sync/store"
	"github.com/irisnet/irishub-sync/logger"
	"strconv"
)

func HandleTxMsg(msgData sdk.Msg, docTx *document.CommonTx) (*document.CommonTx, bool) {
	ok := true
	switch msgData.Type() {
	case new(types.MsgSubmitProposal).Type():
		txMsg := DocTxMsgSubmitProposal{}
		txMsg.BuildMsg(msgData)
		docTx.Msgs = append(docTx.Msgs, document.DocTxMsg{
			Type: txMsg.Type(),
			Msg:  &txMsg,
		})
		docTx.Addrs = append(docTx.Addrs, txMsg.Proposer)
		docTx.Types = append(docTx.Types, txMsg.Type())
		if len(docTx.Msgs) > 1 {
			return docTx, true
		}
		docTx.Type = constant.TxTypeSubmitProposal
		//query proposal_id
		proposalId, amount, err := getProposalIdFromEvents(docTx.Events)
		if err != nil {
			logger.Error("can't get proposal id from tags", logger.String("txHash", docTx.TxHash),
				logger.String("err", err.Error()))
		}
		docTx.ProposalId = proposalId
		docTx.Amount = store.Coins{amount}
		if len(docTx.Signers) > 0 {
			docTx.From = docTx.Signers[0].AddrBech32
		}

	case new(types.MsgDeposit).Type():
		txMsg := DocTxMsgDeposit{}
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
		docTx.Type = constant.TxTypeDeposit
		docTx.ProposalId = txMsg.ProposalID

	case new(types.MsgVote).Type():
		txMsg := DocTxMsgVote{}
		txMsg.BuildMsg(msgData)
		docTx.Msgs = append(docTx.Msgs, document.DocTxMsg{
			Type: txMsg.Type(),
			Msg:  &txMsg,
		})
		docTx.Addrs = append(docTx.Addrs, txMsg.Voter)
		docTx.Types = append(docTx.Types, txMsg.Type())
		if len(docTx.Msgs) > 1 {
			return docTx, true
		}
		docTx.From = txMsg.Voter
		docTx.Amount = []store.Coin{}
		docTx.Type = constant.TxTypeVote
		docTx.ProposalId = txMsg.ProposalID
	default:
		ok = false
	}
	return docTx, ok
}

// get proposalId from tags
func getProposalIdFromEvents(events []document.Event) (uint64, store.Coin, error) {
	//query proposal_id
	//for _, tag := range tags {
	//	key := string(tag.Key)
	//	if key == types.EventGovProposalId {
	//		if proposalId, err := strconv.ParseInt(string(tag.Value), 10, 0); err != nil {
	//			return 0, err
	//		} else {
	//			return uint64(proposalId), nil
	//		}
	//	}
	//}
	var proposalId uint64
	var amount store.Coin
	for _, val := range events {
		if val.Type != types.EventTypeProposalDeposit {
			continue
		}
		for _, attr := range val.Attributes {
			if string(attr.Key) == types.EventGovProposalID {
				if id, err := strconv.ParseInt(string(attr.Value), 10, 0); err == nil {
					proposalId = uint64(id)
				}
			}
			if string(attr.Key) == "amount" && string(attr.Value) != "" {
				value := string(attr.Value)
				amount = types.ParseCoin(value)
			}
		}
	}

	return proposalId, amount, nil
}
