// package for parse tx struct from binary data

package helper

import (
	"encoding/hex"
	"github.com/irisnet/irishub-sync/logger"
	"github.com/irisnet/irishub-sync/store"
	"github.com/irisnet/irishub-sync/store/document"
	itypes "github.com/irisnet/irishub-sync/types"
	imsg "github.com/irisnet/irishub-sync/types/msg"
	"github.com/irisnet/irishub-sync/util/constant"
	"strconv"
	"strings"
	"time"
)

func ParseTx(txBytes itypes.Tx, block *itypes.Block) document.CommonTx {
	var (
		authTx     itypes.StdTx
		methodName = "ParseTx"
		docTx      document.CommonTx
		gasPrice   float64
		actualFee  store.ActualFee
		signers    []document.Signer
		docTxMsgs  []document.DocTxMsg
	)

	cdc := itypes.GetCodec()

	err := cdc.UnmarshalBinaryLengthPrefixed(txBytes, &authTx)
	if err != nil {
		logger.Error(err.Error())
		return docTx
	}

	height := block.Height
	blockTime := block.Time
	txHash := BuildHex(txBytes.Hash())
	fee := itypes.BuildFee(authTx.Fee)
	memo := authTx.Memo

	// get tx signers
	if len(authTx.Signatures) > 0 {
		for _, signature := range authTx.Signatures {
			address := signature.Address()

			signer := document.Signer{}
			signer.AddrHex = address.String()
			if addrBech32, err := ConvertAccountAddrFromHexToBech32(address.Bytes()); err != nil {
				logger.Error("convert account addr from hex to bech32 fail",
					logger.String("addrHex", address.String()), logger.String("err", err.Error()))
			} else {
				signer.AddrBech32 = addrBech32
			}
			signers = append(signers, signer)
		}
	}

	// get tx status, gasUsed, gasPrice and actualFee from tx result
	status, result, err := QueryTxResult(txBytes.Hash())
	if err != nil {
		logger.Error("get txResult err", logger.String("method", methodName), logger.String("err", err.Error()))
	}
	log := result.Log
	gasUsed := Min(result.GasUsed, fee.Gas)
	if len(fee.Amount) > 0 {
		gasPrice = fee.Amount[0].Amount / float64(fee.Gas)
		actualFee = store.ActualFee{
			Denom:  fee.Amount[0].Denom,
			Amount: float64(gasUsed) * gasPrice,
		}
	} else {
		gasPrice = 0
		actualFee = store.ActualFee{}
	}

	msgs := authTx.GetMsgs()
	if len(msgs) <= 0 {
		logger.Error("can't get msgs", logger.String("method", methodName))
		return docTx
	}
	msg := msgs[0]

	docTx = document.CommonTx{
		Height:    height,
		Time:      blockTime,
		TxHash:    txHash,
		Fee:       fee,
		Memo:      memo,
		Status:    status,
		Code:      result.Code,
		Log:       log,
		GasUsed:   gasUsed,
		GasWanted: result.GasUsed,
		GasPrice:  gasPrice,
		ActualFee: actualFee,
		Events:    parseEvents(result),
		Signers:   signers,
	}

	switch msg.(type) {
	case itypes.MsgTransfer:
		msg := msg.(itypes.MsgTransfer)

		docTx.From = msg.FromAddress.String()
		docTx.To = msg.ToAddress.String()
		docTx.Amount = itypes.ParseCoins(msg.Amount.String())
		docTx.Type = constant.TxTypeTransfer
		txMsg := imsg.DocTxMsgSend{}
		txMsg.BuildMsg(msg)
		docTx.Msgs = append(docTxMsgs, document.DocTxMsg{
			Type: txMsg.Type(),
			Msg:  &txMsg,
		})
		return docTx

	case itypes.MsgStakeCreate:
		msg := msg.(itypes.MsgStakeCreate)

		docTx.From = msg.DelegatorAddress.String()
		docTx.To = msg.ValidatorAddress.String()
		docTx.Amount = []store.Coin{itypes.ParseCoin(msg.Value.String())}
		docTx.Type = constant.TxTypeStakeCreateValidator
		txMsg := imsg.DocTxMsgStakeCreate{}
		txMsg.BuildMsg(msg)
		docTx.Msgs = append(docTxMsgs, document.DocTxMsg{
			Type: txMsg.Type(),
			Msg:  &txMsg,
		})
		return docTx
	case itypes.MsgStakeEdit:
		msg := msg.(itypes.MsgStakeEdit)

		docTx.From = msg.ValidatorAddress.String()
		docTx.To = ""
		docTx.Amount = []store.Coin{}
		docTx.Type = constant.TxTypeStakeEditValidator
		txMsg := imsg.DocTxMsgStakeEdit{}
		txMsg.BuildMsg(msg)
		docTx.Msgs = append(docTxMsgs, document.DocTxMsg{
			Type: txMsg.Type(),
			Msg:  &txMsg,
		})
		return docTx
	case itypes.MsgStakeDelegate:
		msg := msg.(itypes.MsgStakeDelegate)

		docTx.From = msg.DelegatorAddress.String()
		docTx.To = msg.ValidatorAddress.String()
		docTx.Amount = []store.Coin{itypes.ParseCoin(msg.Amount.String())}
		docTx.Type = constant.TxTypeStakeDelegate
		txMsg := imsg.DocTxMsgDelegate{}
		txMsg.BuildMsg(msg)
		docTx.Msgs = append(docTxMsgs, document.DocTxMsg{
			Type: txMsg.Type(),
			Msg:  &txMsg,
		})

		return docTx
	case itypes.MsgStakeBeginUnbonding:
		msg := msg.(itypes.MsgStakeBeginUnbonding)

		shares := ParseFloat(msg.Amount.String())
		docTx.From = msg.DelegatorAddress.String()
		docTx.To = msg.ValidatorAddress.String()

		coin := store.Coin{
			Amount: shares,
		}
		docTx.Amount = []store.Coin{coin}
		docTx.Type = constant.TxTypeStakeBeginUnbonding
		txMsg := imsg.DocTxMsgBeginUnbonding{}
		txMsg.BuildMsg(msg)
		docTx.Msgs = append(docTxMsgs, document.DocTxMsg{
			Type: txMsg.Type(),
			Msg:  &txMsg,
		})
		return docTx
	case itypes.MsgBeginRedelegate:
		msg := msg.(itypes.MsgBeginRedelegate)

		shares := ParseFloat(msg.Amount.String())
		docTx.From = msg.ValidatorSrcAddress.String()
		docTx.To = msg.ValidatorDstAddress.String()
		coin := store.Coin{
			Amount: shares,
		}
		docTx.Amount = []store.Coin{coin}
		docTx.Type = constant.TxTypeBeginRedelegate
		txMsg := imsg.DocTxMsgBeginRedelegate{}
		txMsg.BuildMsg(msg)
		docTx.Msgs = append(docTxMsgs, document.DocTxMsg{
			Type: txMsg.Type(),
			Msg:  &txMsg,
		})
		return docTx
	case itypes.MsgUnjail:
		msg := msg.(itypes.MsgUnjail)

		docTx.From = msg.ValidatorAddr.String()
		docTx.Type = constant.TxTypeUnjail
		txMsg := imsg.DocTxMsgUnjail{}
		txMsg.BuildMsg(msg)
		docTx.Msgs = append(docTxMsgs, document.DocTxMsg{
			Type: txMsg.Type(),
			Msg:  &txMsg,
		})
	case itypes.MsgSetWithdrawAddress:
		msg := msg.(itypes.MsgSetWithdrawAddress)

		docTx.From = msg.DelegatorAddress.String()
		docTx.To = msg.WithdrawAddress.String()
		docTx.Type = constant.TxTypeSetWithdrawAddress
		txMsg := imsg.DocTxMsgSetWithdrawAddress{}
		txMsg.BuildMsg(msg)
		docTx.Msgs = append(docTxMsgs, document.DocTxMsg{
			Type: txMsg.Type(),
			Msg:  &txMsg,
		})
	case itypes.MsgWithdrawDelegatorReward:
		msg := msg.(itypes.MsgWithdrawDelegatorReward)

		docTx.From = msg.DelegatorAddress.String()
		docTx.To = msg.ValidatorAddress.String()
		docTx.Type = constant.TxTypeWithdrawDelegatorReward
		txMsg := imsg.DocTxMsgWithdrawDelegatorReward{}
		txMsg.BuildMsg(msg)
		docTx.Msgs = append(docTxMsgs, document.DocTxMsg{
			Type: txMsg.Type(),
			Msg:  &txMsg,
		})

	case itypes.MsgFundCommunityPool:
		msg := msg.(itypes.MsgFundCommunityPool)

		docTx.From = msg.Depositor.String()
		docTx.Amount = itypes.ParseCoins(msg.Amount.String())
		docTx.Type = constant.TxTypeMsgFundCommunityPool
		txMsg := imsg.DocTxMsgFundCommunityPool{}
		txMsg.BuildMsg(msg)
		docTx.Msgs = append(docTxMsgs, document.DocTxMsg{
			Type: txMsg.Type(),
			Msg:  &txMsg,
		})
	case itypes.MsgWithdrawValidatorCommission:
		msg := msg.(itypes.MsgWithdrawValidatorCommission)

		docTx.From = msg.ValidatorAddress.String()
		docTx.Type = constant.TxTypeMsgWithdrawValidatorCommission
		txMsg := imsg.DocTxMsgWithdrawValidatorCommission{}
		txMsg.BuildMsg(msg)
		docTx.Msgs = append(docTxMsgs, document.DocTxMsg{
			Type: txMsg.Type(),
			Msg:  &txMsg,
		})

	case itypes.MsgSubmitProposal:
		msg := msg.(itypes.MsgSubmitProposal)

		docTx.From = msg.Proposer.String()
		docTx.To = ""
		docTx.Amount = itypes.ParseCoins(msg.InitialDeposit.String())
		docTx.Type = constant.TxTypeSubmitProposal
		txMsg := imsg.DocTxMsgSubmitProposal{}
		txMsg.BuildMsg(msg)
		docTx.Msgs = append(docTxMsgs, document.DocTxMsg{
			Type: txMsg.Type(),
			Msg:  &txMsg,
		})

		//query proposal_id
		proposalId, err := getProposalIdFromTags(result.Tags)
		if err != nil {
			logger.Error("can't get proposal id from tags", logger.String("txHash", docTx.TxHash),
				logger.String("err", err.Error()))
		}
		docTx.ProposalId = proposalId

		return docTx
		//case itypes.MsgSubmitSoftwareUpgradeProposal:
		//	msg := msg.(itypes.MsgSubmitSoftwareUpgradeProposal)
		//
		//	docTx.From = msg.Proposer.String()
		//	docTx.To = ""
		//	docTx.Amount = itypes.ParseCoins(msg.InitialDeposit.String())
		//	docTx.Type = constant.TxTypeSubmitProposal
		//	txMsg := imsg.DocTxMsgSubmitSoftwareUpgradeProposal{}
		//	txMsg.BuildMsg(msg)
		//	docTx.Msgs = append(docTxMsgs, document.DocTxMsg{
		//		Type: txMsg.Type(),
		//		Msg:  &txMsg,
		//	})
		//
		//	//query proposal_id
		//	proposalId, err := getProposalIdFromTags(result.Tags)
		//	if err != nil {
		//		logger.Error("can't get proposal id from tags", logger.String("txHash", docTx.TxHash),
		//			logger.String("err", err.Error()))
		//	}
		//	docTx.ProposalId = proposalId
		//
		//	return docTx
		//case itypes.MsgSubmitTaxUsageProposal:
		//	msg := msg.(itypes.MsgSubmitTaxUsageProposal)
		//
		//	docTx.From = msg.Proposer.String()
		//	docTx.To = ""
		//	docTx.Amount = itypes.ParseCoins(msg.InitialDeposit.String())
		//	docTx.Type = constant.TxTypeSubmitProposal
		//	txMsg := imsg.DocTxMsgSubmitCommunityTaxUsageProposal{}
		//	txMsg.BuildMsg(msg)
		//	docTx.Msgs = append(docTxMsgs, document.DocTxMsg{
		//		Type: txMsg.Type(),
		//		Msg:  &txMsg,
		//	})
		//
		//	//query proposal_id
		//	proposalId, err := getProposalIdFromTags(result.Tags)
		//	if err != nil {
		//		logger.Error("can't get proposal id from tags", logger.String("txHash", docTx.TxHash),
		//			logger.String("err", err.Error()))
		//	}
		//	docTx.ProposalId = proposalId
		//	return docTx
		//case itypes.MsgSubmitTokenAdditionProposal:
		//	msg := msg.(itypes.MsgSubmitTokenAdditionProposal)
		//
		//	docTx.From = msg.Proposer.String()
		//	docTx.To = ""
		//	docTx.Amount = itypes.ParseCoins(msg.InitialDeposit.String())
		//	docTx.Type = constant.TxTypeSubmitProposal
		//	txMsg := imsg.DocTxMsgSubmitTokenAdditionProposal{}
		//	txMsg.BuildMsg(msg)
		//	docTx.Msgs = append(docTxMsgs, document.DocTxMsg{
		//		Type: txMsg.Type(),
		//		Msg:  &txMsg,
		//	})
		//	//query proposal_id
		//	proposalId, err := getProposalIdFromTags(result.Tags)
		//	if err != nil {
		//		logger.Error("can't get proposal id from tags", logger.String("txHash", docTx.TxHash),
		//			logger.String("err", err.Error()))
		//	}
		//	docTx.ProposalId = proposalId
		//	return docTx
	case itypes.MsgDeposit:
		msg := msg.(itypes.MsgDeposit)

		docTx.From = msg.Depositor.String()
		docTx.Amount = itypes.ParseCoins(msg.Amount.String())
		docTx.Type = constant.TxTypeDeposit
		docTx.ProposalId = msg.ProposalID
		txMsg := imsg.DocTxMsgDeposit{}
		txMsg.BuildMsg(msg)
		docTx.Msgs = append(docTxMsgs, document.DocTxMsg{
			Type: txMsg.Type(),
			Msg:  &txMsg,
		})
		return docTx
	case itypes.MsgVote:
		msg := msg.(itypes.MsgVote)

		docTx.From = msg.Voter.String()
		docTx.Amount = []store.Coin{}
		docTx.Type = constant.TxTypeVote
		docTx.ProposalId = msg.ProposalID
		txMsg := imsg.DocTxMsgVote{}
		txMsg.BuildMsg(msg)
		docTx.Msgs = append(docTxMsgs, document.DocTxMsg{
			Type: txMsg.Type(),
			Msg:  &txMsg,
		})
		return docTx
	case itypes.MsgRequestRandom:
		msg := msg.(itypes.MsgRequestRandom)

		docTx.From = msg.Consumer.String()
		docTx.Amount = []store.Coin{}
		docTx.Type = constant.TxTypeRequestRand
		txMsg := imsg.DocTxMsgRequestRand{}
		txMsg.BuildMsg(msg)
		docTx.Msgs = append(docTxMsgs, document.DocTxMsg{
			Type: txMsg.Type(),
			Msg:  &txMsg,
		})
		return docTx
	case itypes.AssetIssueToken:
		msg := msg.(itypes.AssetIssueToken)

		docTx.From = msg.Owner.String()
		docTx.Type = constant.TxTypeAssetIssueToken
		txMsg := imsg.DocTxMsgIssueToken{}
		txMsg.BuildMsg(msg)
		docTx.Msgs = append(docTxMsgs, document.DocTxMsg{
			Type: txMsg.Type(),
			Msg:  &txMsg,
		})

		return docTx
	case itypes.AssetEditToken:
		msg := msg.(itypes.AssetEditToken)

		docTx.From = msg.Owner.String()
		docTx.Type = constant.TxTypeAssetEditToken
		txMsg := imsg.DocTxMsgEditToken{}
		txMsg.BuildMsg(msg)
		docTx.Msgs = append(docTxMsgs, document.DocTxMsg{
			Type: txMsg.Type(),
			Msg:  &txMsg,
		})

		return docTx
	case itypes.AssetMintToken:
		msg := msg.(itypes.AssetMintToken)

		docTx.From = msg.Owner.String()
		docTx.To = msg.To.String()
		docTx.Type = constant.TxTypeAssetMintToken
		txMsg := imsg.DocTxMsgMintToken{}
		txMsg.BuildMsg(msg)
		docTx.Msgs = append(docTxMsgs, document.DocTxMsg{
			Type: txMsg.Type(),
			Msg:  &txMsg,
		})

		return docTx
	case itypes.AssetTransferTokenOwner:
		msg := msg.(itypes.AssetTransferTokenOwner)

		docTx.From = msg.SrcOwner.String()
		docTx.To = msg.DstOwner.String()
		docTx.Type = constant.TxTypeAssetTransferTokenOwner
		txMsg := imsg.DocTxMsgTransferTokenOwner{}
		txMsg.BuildMsg(msg)
		docTx.Msgs = append(docTxMsgs, document.DocTxMsg{
			Type: txMsg.Type(),
			Msg:  &txMsg,
		})

		return docTx
		//case itypes.AssetCreateGateway:
		//	msg := msg.(itypes.AssetCreateGateway)
		//
		//	docTx.From = msg.Owner.String()
		//	docTx.Type = constant.TxTypeAssetCreateGateway
		//	txMsg := imsg.DocTxMsgCreateGateway{}
		//	txMsg.BuildMsg(msg)
		//	docTx.Msgs = append(docTxMsgs, document.DocTxMsg{
		//		Type: txMsg.Type(),
		//		Msg:  &txMsg,
		//	})
		//
		//	return docTx
		//case itypes.AssetEditGateWay:
		//	msg := msg.(itypes.AssetEditGateWay)
		//
		//	docTx.From = msg.Owner.String()
		//	docTx.Type = constant.TxTypeAssetEditGateway
		//	txMsg := imsg.DocTxMsgEditGateway{}
		//	txMsg.BuildMsg(msg)
		//	docTx.Msgs = append(docTxMsgs, document.DocTxMsg{
		//		Type: txMsg.Type(),
		//		Msg:  &txMsg,
		//	})
		//
		//	return docTx
		//case itypes.AssetTransferGatewayOwner:
		//	msg := msg.(itypes.AssetTransferGatewayOwner)
		//
		//	docTx.From = msg.Owner.String()
		//	docTx.To = msg.To.String()
		//	docTx.Type = constant.TxTypeAssetTransferGatewayOwner
		//	txMsg := imsg.DocTxMsgTransferGatewayOwner{}
		//	txMsg.BuildMsg(msg)
		//	docTx.Msgs = append(docTxMsgs, document.DocTxMsg{
		//		Type: txMsg.Type(),
		//		Msg:  &txMsg,
		//	})
		//	return docTx

	case itypes.MsgAddProfiler:
		msg := msg.(itypes.MsgAddProfiler)

		docTx.From = msg.AddGuardian.AddedBy.String()
		docTx.To = msg.AddGuardian.Address.String()
		docTx.Type = constant.TxTypeAddProfiler
		txMsg := imsg.DocTxMsgAddProfiler{}
		txMsg.BuildMsg(msg)
		docTx.Msgs = append(docTxMsgs, document.DocTxMsg{
			Type: txMsg.Type(),
			Msg:  &txMsg,
		})
		return docTx

	case itypes.MsgAddTrustee:
		msg := msg.(itypes.MsgAddTrustee)

		docTx.From = msg.AddGuardian.AddedBy.String()
		docTx.To = msg.AddGuardian.Address.String()
		docTx.Type = constant.TxTypeAddTrustee
		txMsg := imsg.DocTxMsgAddTrustee{}
		txMsg.BuildMsg(msg)
		docTx.Msgs = append(docTxMsgs, document.DocTxMsg{
			Type: txMsg.Type(),
			Msg:  &txMsg,
		})
		return docTx

	case itypes.MsgDeleteTrustee:
		msg := msg.(itypes.MsgDeleteTrustee)

		docTx.From = msg.DeleteGuardian.DeletedBy.String()
		docTx.To = msg.DeleteGuardian.Address.String()
		docTx.Type = constant.TxTypeDeleteTrustee
		txMsg := imsg.DocTxMsgDeleteTrustee{}
		txMsg.BuildMsg(msg)
		docTx.Msgs = append(docTxMsgs, document.DocTxMsg{
			Type: txMsg.Type(),
			Msg:  &txMsg,
		})
		return docTx

	case itypes.MsgDeleteProfiler:
		msg := msg.(itypes.MsgDeleteProfiler)

		docTx.From = msg.DeleteGuardian.DeletedBy.String()
		docTx.To = msg.DeleteGuardian.Address.String()
		docTx.Type = constant.TxTypeDeleteProfiler
		txMsg := imsg.DocTxMsgDeleteProfiler{}
		txMsg.BuildMsg(msg)
		docTx.Msgs = append(docTxMsgs, document.DocTxMsg{
			Type: txMsg.Type(),
			Msg:  &txMsg,
		})
		return docTx

	case itypes.MsgCreateHTLC:
		msg := msg.(itypes.MsgCreateHTLC)

		docTx.From = msg.Sender.String()
		docTx.To = msg.To.String()
		docTx.Amount = itypes.ParseCoins(msg.Amount.String())
		docTx.Type = constant.TxTypeCreateHTLC
		txMsg := imsg.DocTxMsgCreateHTLC{}
		txMsg.BuildMsg(msg)
		docTx.Msgs = append(docTxMsgs, document.DocTxMsg{
			Type: txMsg.Type(),
			Msg:  &txMsg,
		})
		return docTx
	case itypes.MsgClaimHTLC:
		msg := msg.(itypes.MsgClaimHTLC)

		docTx.From = msg.Sender.String()
		docTx.To = ""
		docTx.Type = constant.TxTypeClaimHTLC
		txMsg := imsg.DocTxMsgClaimHTLC{}
		txMsg.BuildMsg(msg)
		docTx.Msgs = append(docTxMsgs, document.DocTxMsg{
			Type: txMsg.Type(),
			Msg:  &txMsg,
		})
		return docTx
	case itypes.MsgRefundHTLC:
		msg := msg.(itypes.MsgRefundHTLC)

		docTx.From = msg.Sender.String()
		docTx.To = ""
		docTx.Type = constant.TxTypeRefundHTLC
		txMsg := imsg.DocTxMsgRefundHTLC{}
		txMsg.BuildMsg(msg)
		docTx.Msgs = append(docTxMsgs, document.DocTxMsg{
			Type: txMsg.Type(),
			Msg:  &txMsg,
		})
		return docTx
	case itypes.MsgAddLiquidity:
		msg := msg.(itypes.MsgAddLiquidity)

		docTx.From = msg.Sender.String()
		docTx.To = ""
		docTx.Amount = itypes.ParseCoins(msg.MaxToken.String())
		docTx.Type = constant.TxTypeAddLiquidity
		txMsg := imsg.DocTxMsgAddLiquidity{}
		txMsg.BuildMsg(msg)
		docTx.Msgs = append(docTxMsgs, document.DocTxMsg{
			Type: txMsg.Type(),
			Msg:  &txMsg,
		})
		return docTx
	case itypes.MsgRemoveLiquidity:
		msg := msg.(itypes.MsgRemoveLiquidity)

		docTx.From = msg.Sender.String()
		docTx.To = ""
		docTx.Amount = itypes.ParseCoins(msg.WithdrawLiquidity.String())
		docTx.Type = constant.TxTypeRemoveLiquidity
		txMsg := imsg.DocTxMsgRemoveLiquidity{}
		txMsg.BuildMsg(msg)
		docTx.Msgs = append(docTxMsgs, document.DocTxMsg{
			Type: txMsg.Type(),
			Msg:  &txMsg,
		})
		return docTx
	case itypes.MsgSwapOrder:
		msg := msg.(itypes.MsgSwapOrder)

		docTx.From = msg.Input.Address.String()
		docTx.To = msg.Output.Address.String()
		docTx.Amount = itypes.ParseCoins(msg.Input.Coin.String())
		docTx.Type = constant.TxTypeSwapOrder
		txMsg := imsg.DocTxMsgSwapOrder{}
		txMsg.BuildMsg(msg)
		docTx.Msgs = append(docTxMsgs, document.DocTxMsg{
			Type: txMsg.Type(),
			Msg:  &txMsg,
		})
		return docTx

	default:
		logger.Warn("unknown msg type")
	}

	return docTx
}

func parseEvents(result itypes.ResponseDeliverTx) []document.Event {

	var events []document.Event
	for _, val := range result.GetEvents() {
		one := document.Event{
			Type: val.Type,
		}
		one.Attributes = make(map[string]string, len(val.Attributes))
		for _, attr := range val.Attributes {
			one.Attributes[string(attr.Key)] = string(attr.Value)
		}
		events = append(events, one)
	}

	return events
}

// get proposalId from tags
func getProposalIdFromTags(tags []itypes.TmKVPair) (uint64, error) {
	//query proposal_id
	for _, tag := range tags {
		key := string(tag.Key)
		if key == itypes.TagGovProposalID {
			if proposalId, err := strconv.ParseInt(string(tag.Value), 10, 0); err != nil {
				return 0, err
			} else {
				return uint64(proposalId), nil
			}
		}
	}
	return 0, nil
}

func BuildHex(bytes []byte) string {
	return strings.ToUpper(hex.EncodeToString(bytes))
}

// get tx status and log by query txHash
func QueryTxResult(txHash []byte) (string, itypes.ResponseDeliverTx, error) {
	var resDeliverTx itypes.ResponseDeliverTx
	status := document.TxStatusSuccess

	client := GetClient()
	defer client.Release()

	res, err := client.Tx(txHash, false)
	if err != nil {
		// try again
		time.Sleep(time.Duration(1) * time.Second)
		if res, err := client.Tx(txHash, false); err != nil {
			return "unknown", resDeliverTx, err
		} else {
			resDeliverTx = res.TxResult
		}
	} else {
		resDeliverTx = res.TxResult
	}

	if resDeliverTx.Code != 0 {
		status = document.TxStatusFail
	}

	return status, resDeliverTx, nil
}
