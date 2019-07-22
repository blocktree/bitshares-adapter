package bitshares

import (
	"testing"

	"github.com/blocktree/bitshares-adapter/types"
	"github.com/blocktree/openwallet/log"
)

func TestWalletClient_GetBlockchainInfo(t *testing.T) {
	b, err := tw.Api.GetBlockchainInfo()
	if err != nil {
		t.Errorf("GetBlockchainInfo failed unexpected error: %v\n", err)
	} else {
		log.Infof("GetBlockchainInfo info: %+v\n", b)
	}
}

func TestWalletClient_GetBlockByHeight(t *testing.T) {
	block, err := tw.Api.GetBlockByHeight(161025)
	if err != nil {
		t.Errorf("GetBlockByHeight failed unexpected error: %v\n", err)
	} else {
		log.Infof("GetBlockByHeight info: %+v", block)
	}
}

func TestWalletClient_GetTransaction(t *testing.T) {
	tx, err := tw.Api.GetTransaction(1545399, 0)
	if err != nil {
		t.Errorf("GetTransaction failed unexpected error: %v\n", err)
	} else {
		log.Infof("GetTransaction info: %+v", tx)
	}
}

func TestWalletClient_GetAssetsBalance(t *testing.T) {
	balances, err := tw.Api.GetAssetsBalance(types.MustParseObjectID("1.2.814225"), types.MustParseObjectID("1.3.0"))
	if err != nil {
		t.Errorf("Balances failed unexpected error: %v\n", err)
	} else {
		log.Infof("Balances info: %+v", balances)
	}
}
