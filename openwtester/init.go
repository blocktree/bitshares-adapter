package openwtester

import (
	"github.com/blocktree/bitshares-adapter/bitshares"
	"github.com/blocktree/openwallet/log"
	"github.com/blocktree/openwallet/openw"
)

func init() {
	//注册钱包管理工具
	log.Notice("Wallet Manager Load Successfully.")
	// openw.RegAssets(bitshares.Symbol, bitshares.NewWalletManager(nil))

	cache := bitshares.NewCacheManager()

	openw.RegAssets(bitshares.Symbol, bitshares.NewWalletManager(&cache))
}
