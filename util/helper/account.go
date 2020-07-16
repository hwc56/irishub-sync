// This package is used for Query balance of account

package helper

import (
	"github.com/irisnet/irishub-sync/types"
	"github.com/tendermint/tendermint/libs/bech32"
)


// convert account address from hex to bech32
func ConvertAccountAddrFromHexToBech32(address []byte) (string, error) {
	return bech32.ConvertAndEncode(types.Bech32AccountAddrPrefix, address)
}
