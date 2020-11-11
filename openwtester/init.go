package openwtester

import (
	"github.com/blocktree/bitshares-adapter/bitshares"
	"github.com/blocktree/openwallet/v2/log"
	"github.com/blocktree/openwallet/v2/openw"
)

func init() {
	//注册钱包管理工具
	log.Notice("Wallet Manager Load Successfully.")
	cache := bitshares.NewCacheManager()

	openw.RegAssets(bitshares.Symbol, bitshares.NewWalletManager(&cache))
}
