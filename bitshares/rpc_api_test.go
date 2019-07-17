package bitshares

import (
	"testing"

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
	tx, err := tw.Api.GetTransaction(1545399, 1)
	if err != nil {
		t.Errorf("GetTransaction failed unexpected error: %v\n", err)
	} else {
		log.Infof("GetTransaction info: %+v", tx)
	}
}
