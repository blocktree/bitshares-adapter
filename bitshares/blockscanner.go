/*
 * Copyright 2018 The OpenWallet Authors
 * This file is part of the OpenWallet library.
 *
 * The OpenWallet library is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The OpenWallet library is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU Lesser General Public License for more details.
 */

package bitshares

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/blocktree/bitshares-adapter/types"

	"github.com/blocktree/openwallet/common"
	"github.com/blocktree/openwallet/log"
	"github.com/blocktree/openwallet/openwallet"
)

const (
	blockchainBucket = "blockchain" // blockchain dataset
	//periodOfTask      = 5 * time.Second // task interval
	maxExtractingSize = 10 // thread count
)

//BtsBlockScanner BTS block scanner
type BtsBlockScanner struct {
	*openwallet.BlockScannerBase

	CurrentBlockHeight   uint64         //当前区块高度
	extractingCH         chan struct{}  //扫描工作令牌
	wm                   *WalletManager //钱包管理者
	IsScanMemPool        bool           //是否扫描交易池
	RescanLastBlockCount uint64         //重扫上N个区块数量
}

//ExtractResult extract result
type ExtractResult struct {
	extractData map[string][]*openwallet.TxExtractData
	TxID        string
	BlockHash   string
	BlockHeight uint64
	BlockTime   int64
	Success     bool
}

//SaveResult result
type SaveResult struct {
	TxID        string
	BlockHeight uint64
	Success     bool
}

// NewBlockScanner create a block scanner
func NewBlockScanner(wm *WalletManager) *BtsBlockScanner {
	bs := BtsBlockScanner{
		BlockScannerBase: openwallet.NewBlockScannerBase(),
	}

	bs.extractingCH = make(chan struct{}, maxExtractingSize)
	bs.wm = wm
	bs.IsScanMemPool = true
	bs.RescanLastBlockCount = 0

	// set task
	bs.SetTask(bs.ScanBlockTask)

	return &bs
}

// ScanBlockTask scan block task
func (bs *BtsBlockScanner) ScanBlockTask() {

	var (
		currentHeight uint32
		currentHash   string
	)

	// get local block header
	currentHeight, currentHash, err := bs.GetLocalBlockHead()

	if err != nil {
		bs.wm.Log.Std.Error("", err)
	}

	if currentHeight == 0 {
		bs.wm.Log.Std.Info("No records found in local, get current block as the local!")

		headBlock, err := bs.GetGlobalHeadBlock()
		if err != nil {
			bs.wm.Log.Std.Info("get head block error, err=%v", err)
		}

		currentHash = headBlock.Previous
		currentHeight = uint32(headBlock.Height - 1)
	}

	for {
		if !bs.Scanning {
			// stop scan
			return
		}

		infoResp, err := bs.GetChainInfo()
		if err != nil {
			bs.wm.Log.Errorf("get chain info failed, err=%v", err)
			break
		}

		maxBlockHeight := infoResp.HeadBlockNum

		bs.wm.Log.Info("current block height:", currentHeight, " maxBlockHeight:", maxBlockHeight)
		if uint64(currentHeight) == maxBlockHeight-1 {
			bs.wm.Log.Std.Info("block scanner has scanned full chain data. Current height %d", maxBlockHeight)
			break
		}

		// next block
		currentHeight = currentHeight + 1

		bs.wm.Log.Std.Info("block scanner scanning height: %d ...", currentHeight)
		block, err := bs.wm.Api.GetBlockByHeight(currentHeight)

		if err != nil {
			bs.wm.Log.Std.Info("block scanner can not get new block data by rpc; unexpected error: %v", err)
			break
		}

		if currentHash != block.Previous {
			bs.wm.Log.Std.Info("block has been fork on height: %d.", currentHeight)
			bs.wm.Log.Std.Info("block height: %d local hash = %s ", currentHeight-1, currentHash)
			bs.wm.Log.Std.Info("block height: %d mainnet hash = %s ", currentHeight-1, block.Previous)
			bs.wm.Log.Std.Info("delete recharge records on block height: %d.", currentHeight-1)

			// get local fork bolck
			forkBlock, _ := bs.GetLocalBlock(currentHeight - 1)
			// delete last unscan block
			bs.DeleteUnscanRecord(currentHeight - 1)
			currentHeight = currentHeight - 2 // scan back to last 2 block
			if currentHeight <= 0 {
				currentHeight = 1
			}
			localBlock, err := bs.GetLocalBlock(currentHeight)
			if err != nil {
				bs.wm.Log.Std.Error("block scanner can not get local block; unexpected error: %v", err)
				//get block from rpc
				bs.wm.Log.Info("block scanner prev block height:", currentHeight)
				curBlock, err := bs.wm.Api.GetBlockByHeight(currentHeight)
				if err != nil {
					bs.wm.Log.Std.Error("block scanner can not get prev block by rpc; unexpected error: %v", err)
					break
				}
				currentHash = curBlock.BlockID
			} else {
				//重置当前区块的hash
				currentHash = localBlock.BlockID
			}
			bs.wm.Log.Std.Info("rescan block on height: %d, hash: %s .", currentHeight, currentHash)

			//重新记录一个新扫描起点
			bs.SaveLocalBlockHead(currentHeight, currentHash)

			if forkBlock != nil {
				//通知分叉区块给观测者，异步处理
				bs.forkBlockNotify(forkBlock)
			}

		} else {
			currentHash = block.BlockID
			err := bs.BatchExtractTransactions(uint64(currentHeight), currentHash, block.Timestamp.Unix(), block.Transactions, block.TransactionIDs)
			if err != nil {
				bs.wm.Log.Std.Error("block scanner ran BatchExtractTransactions occured unexpected error: %v", err)
			}

			//保存本地新高度
			bs.SaveLocalBlockHead(currentHeight, currentHash)
			bs.SaveLocalBlock(block)
			//通知新区块给观测者，异步处理
			bs.newBlockNotify(block)
		}
	}

	//重扫失败区块
	bs.RescanFailedRecord()

}

//newBlockNotify 获得新区块后，通知给观测者
func (bs *BtsBlockScanner) forkBlockNotify(block *Block) {
	header := ParseHeader(block)
	header.Fork = true
	bs.NewBlockNotify(header)
}

//newBlockNotify 获得新区块后，通知给观测者
func (bs *BtsBlockScanner) newBlockNotify(block *Block) {
	header := ParseHeader(block)
	bs.NewBlockNotify(header)
}

// BatchExtractTransactions 批量提取交易单
func (bs *BtsBlockScanner) BatchExtractTransactions(blockHeight uint64, blockHash string, blockTime int64, transactions []*types.Transaction, txIDs []string) error {

	var (
		quit       = make(chan struct{})
		done       = 0 //完成标记
		failed     = 0
		shouldDone = len(transactions) //需要完成的总数
	)

	if len(transactions) == 0 {
		return nil
	}

	bs.wm.Log.Std.Info("block scanner ready extract transactions total: %d ", len(transactions))

	//生产通道
	producer := make(chan ExtractResult)
	defer close(producer)

	//消费通道
	worker := make(chan ExtractResult)
	defer close(worker)

	//保存工作
	saveWork := func(height uint64, result chan ExtractResult) {
		//回收创建的地址
		for gets := range result {

			if gets.Success {
				notifyErr := bs.newExtractDataNotify(height, gets.extractData)
				if notifyErr != nil {
					failed++ //标记保存失败数
					bs.wm.Log.Std.Info("newExtractDataNotify unexpected error: %v", notifyErr)
				}
			} else {
				//记录未扫区块
				unscanRecord := NewUnscanRecord(height, "", "")
				bs.SaveUnscanRecord(unscanRecord)
				failed++ //标记保存失败数
			}
			//累计完成的线程数
			done++
			if done == shouldDone {
				close(quit) //关闭通道，等于给通道传入nil
			}
		}
	}

	//提取工作
	extractWork := func(eblockHeight uint64, eBlockHash string, eBlockTime int64, mTransactions []*types.Transaction, mtxIDs []string, eProducer chan ExtractResult) {
		for id, tx := range mTransactions {
			bs.extractingCH <- struct{}{}

			if len(mtxIDs) > id {
				tx.TransactionID = mtxIDs[id]
			}

			go func(mBlockHeight uint64, mTx *types.Transaction, end chan struct{}, mProducer chan<- ExtractResult) {
				//导出提出的交易
				mProducer <- bs.ExtractTransaction(mBlockHeight, eBlockHash, eBlockTime, mTx, bs.ScanTargetFunc)
				//释放
				<-end

			}(eblockHeight, tx, bs.extractingCH, eProducer)
		}
	}
	/*	开启导出的线程	*/

	//独立线程运行消费
	go saveWork(blockHeight, worker)

	//独立线程运行生产
	go extractWork(blockHeight, blockHash, blockTime, transactions, txIDs, producer)

	//以下使用生产消费模式
	bs.extractRuntime(producer, worker, quit)

	if failed > 0 {
		return fmt.Errorf("block scanner saveWork failed")
	}

	return nil
}

//extractRuntime 提取运行时
func (bs *BtsBlockScanner) extractRuntime(producer chan ExtractResult, worker chan ExtractResult, quit chan struct{}) {

	var (
		values = make([]ExtractResult, 0)
	)

	for {
		var activeWorker chan<- ExtractResult
		var activeValue ExtractResult
		//当数据队列有数据时，释放顶部，传输给消费者
		if len(values) > 0 {
			activeWorker = worker
			activeValue = values[0]
		}
		select {
		//生成者不断生成数据，插入到数据队列尾部
		case pa := <-producer:
			values = append(values, pa)
		case <-quit:
			//退出
			return
		case activeWorker <- activeValue:
			values = values[1:]
		}
	}
	//return
}

// ExtractTransaction 提取交易单
func (bs *BtsBlockScanner) ExtractTransaction(blockHeight uint64, blockHash string, blockTime int64, transaction *types.Transaction, scanTargetFunc openwallet.BlockScanTargetFunc) ExtractResult {
	var (
		success = true
		result  = ExtractResult{
			BlockHash:   blockHash,
			BlockHeight: blockHeight,
			TxID:        transaction.TransactionID,
			extractData: make(map[string][]*openwallet.TxExtractData),
			BlockTime:   blockTime,
		}
	)

	if len(transaction.Operations) == 0 {
		bs.wm.Log.Std.Debug("transaction does not have operation: (sig) %s", transaction.Signatures)
		return ExtractResult{Success: true}
	}

	for _, operation := range transaction.Operations {

		if transferOperation, ok := operation.(*types.TransferOperation); ok {

			txID := transaction.TransactionID
			if len(txID) == 0 {
				txID, err := bs.wm.Api.GetTransactionID(transaction)
				bs.wm.Log.Std.Debug("tx: %v", txID)

				if err != nil || len(txID) == 0 {
					bs.wm.Log.Std.Error("cannot get txid, block: %v %s \n%v", blockHeight, transaction.Signatures, err)
					return ExtractResult{Success: false}
				}
			}
			result.TxID = txID

			if scanTargetFunc == nil {
				bs.wm.Log.Std.Error("scanTargetFunc is not configurated")
				return ExtractResult{Success: false}
			}

			accounts, err := bs.wm.Api.GetAccounts(transferOperation.From.String(), transferOperation.To.String())
			if len(accounts) != 2 {
				bs.wm.Log.Std.Error("cannot get accounts, block: %v %s \n%v", blockHeight, txID, err)
				return ExtractResult{Success: false}
			}
			from := accounts[0]
			to := accounts[1]

			//订阅地址为交易单中的发送者
			accountID1, ok1 := scanTargetFunc(openwallet.ScanTarget{Alias: from.Name, Symbol: bs.wm.Symbol(), BalanceModelType: openwallet.BalanceModelTypeAccount})
			//订阅地址为交易单中的接收者
			accountID2, ok2 := scanTargetFunc(openwallet.ScanTarget{Alias: to.Name, Symbol: bs.wm.Symbol(), BalanceModelType: openwallet.BalanceModelTypeAccount})
			if accountID1 == accountID2 && len(accountID1) > 0 && len(accountID2) > 0 {
				bs.InitExtractResult(accountID1, transferOperation, &result, 0)
			} else {
				if ok1 {
					bs.InitExtractResult(accountID1, transferOperation, &result, 1)
				}

				if ok2 {
					bs.InitExtractResult(accountID2, transferOperation, &result, 2)
				}
			}

		}
	}
	result.Success = success
	return result

}

//InitExtractResult optType = 0: 输入输出提取，1: 输入提取，2：输出提取
func (bs *BtsBlockScanner) InitExtractResult(sourceKey string, operation *types.TransferOperation, result *ExtractResult, optType int64) {

	txExtractDataArray := result.extractData[sourceKey]
	if txExtractDataArray == nil {
		txExtractDataArray = make([]*openwallet.TxExtractData, 0)
	}

	txExtractData := &openwallet.TxExtractData{}

	status := "1"
	reason := ""

	token := operation.Amount.AssetID.String()
	amount := common.NewString(operation.Amount.Amount).String()

	contractID := openwallet.GenContractID(bs.wm.Symbol(), token)
	coin := openwallet.Coin{
		Symbol:     bs.wm.Symbol(),
		IsContract: true,
		ContractID: contractID,
	}

	coin.Contract = openwallet.SmartContract{
		Symbol:     bs.wm.Symbol(),
		ContractID: contractID,
		Address:    token,
		Token:      token,
	}

	accounts, err := bs.wm.Api.GetAccounts(operation.From.String(), operation.To.String())
	if len(accounts) != 2 {
		bs.wm.Log.Std.Error("cannot get accounts, %s %s \n %v", operation.From.String(), operation.To.String(), err)
		return
	}
	from := accounts[0]
	to := accounts[1]

	transx := &openwallet.Transaction{
		Fees:        "0",
		Coin:        coin,
		BlockHash:   result.BlockHash,
		BlockHeight: result.BlockHeight,
		TxID:        result.TxID,
		// Decimal:     0,
		Amount:      amount,
		ConfirmTime: result.BlockTime,
		From:        []string{from.Name + ":" + amount},
		To:          []string{to.Name + ":" + amount},
		IsMemo:      true,
		Status:      status,
		Reason:      reason,
		TxType:      0,
	}

	memo, _ := json.Marshal(operation.Memo)
	transx.SetExtParam("memo", memo)

	wxID := openwallet.GenTransactionWxID(transx)
	transx.WxID = wxID

	txExtractData.Transaction = transx
	if optType == 0 {
		bs.extractTxInput(operation, txExtractData)
		bs.extractTxOutput(operation, txExtractData)
	} else if optType == 1 {
		bs.extractTxInput(operation, txExtractData)
	} else if optType == 2 {
		bs.extractTxOutput(operation, txExtractData)
	}

	txExtractDataArray = append(txExtractDataArray, txExtractData)

	if operation.Fee.AssetID != operation.Amount.AssetID && optType != 2 {
		fee := common.NewString(operation.Fee.Amount).String()
		feeToken := operation.Fee.AssetID.String()

		contractID := openwallet.GenContractID(bs.wm.Symbol(), feeToken)
		feeCoin := openwallet.Coin{
			Symbol:     bs.wm.Symbol(),
			IsContract: true,
			ContractID: contractID,
		}

		feeCoin.Contract = openwallet.SmartContract{
			Symbol:     bs.wm.Symbol(),
			ContractID: contractID,
			Address:    feeToken,
			Token:      feeToken,
		}

		feeTransx := &openwallet.Transaction{
			Fees:        "0",
			Coin:        feeCoin,
			BlockHash:   result.BlockHash,
			BlockHeight: result.BlockHeight,
			TxID:        result.TxID,
			// Decimal:     0,
			Amount:      fee,
			ConfirmTime: result.BlockTime,
			From:        []string{from.Name + ":" + fee},
			IsMemo:      true,
			Status:      status,
			Reason:      reason,
			TxType:      1,
		}

		wxID := openwallet.GenTransactionWxID(feeTransx)
		feeTransx.WxID = wxID

		feeExtractData := &openwallet.TxExtractData{Transaction: feeTransx}
		bs.extractTxInput(operation, feeExtractData)

		txExtractDataArray = append(txExtractDataArray, feeExtractData)
	}

	result.extractData[sourceKey] = txExtractDataArray
}

//extractTxInput 提取交易单输入部分,无需手续费，所以只包含1个TxInput
func (bs *BtsBlockScanner) extractTxInput(operation *types.TransferOperation, txExtractData *openwallet.TxExtractData) {

	tx := txExtractData.Transaction
	coin := openwallet.Coin(tx.Coin)

	accounts, err := bs.wm.Api.GetAccounts(operation.From.String())
	if len(accounts) != 1 {
		bs.wm.Log.Std.Error("cannot get accounts, %s \n %v", operation.From.String(), err)
		return
	}
	from := accounts[0]

	//主网from交易转账信息，第一个TxInput
	txInput := &openwallet.TxInput{}
	txInput.Recharge.Sid = openwallet.GenTxInputSID(tx.TxID, bs.wm.Symbol(), coin.ContractID, uint64(0))
	txInput.Recharge.TxID = tx.TxID
	txInput.Recharge.Address = from.Name
	txInput.Recharge.Coin = coin
	txInput.Recharge.Amount = tx.Amount
	txInput.Recharge.Symbol = coin.Symbol
	//txInput.Recharge.IsMemo = true
	//txInput.Recharge.Memo = data.Memo
	txInput.Recharge.BlockHash = tx.BlockHash
	txInput.Recharge.BlockHeight = tx.BlockHeight
	txInput.Recharge.Index = 0 //账户模型填0
	txInput.Recharge.CreateAt = time.Now().Unix()
	txInput.Recharge.TxType = tx.TxType
	txExtractData.TxInputs = append(txExtractData.TxInputs, txInput)

	if tx.TxType == 0 && operation.Fee.Amount > 0 && operation.Fee.AssetID == operation.Amount.AssetID {
		//手续费也作为一个输出s
		fee := new(big.Int)
		fee.SetUint64(operation.Fee.Amount)
		tmp := *txInput
		feeCharge := &tmp
		feeCharge.Amount = fee.String()
		feeCharge.TxType = 1
		txExtractData.TxInputs = append(txExtractData.TxInputs, feeCharge)
	}

}

//extractTxOutput 提取交易单输入部分,只有一个TxOutPut
func (bs *BtsBlockScanner) extractTxOutput(operation *types.TransferOperation, txExtractData *openwallet.TxExtractData) {

	tx := txExtractData.Transaction
	coin := openwallet.Coin(tx.Coin)

	accounts, err := bs.wm.Api.GetAccounts(operation.To.String())
	if len(accounts) != 1 {
		bs.wm.Log.Std.Error("cannot get accounts, %s \n %v", operation.To.String(), err)
		return
	}
	to := accounts[0]

	//主网to交易转账信息,只有一个TxOutPut
	txOutput := &openwallet.TxOutPut{}
	txOutput.Recharge.Sid = openwallet.GenTxOutPutSID(tx.TxID, bs.wm.Symbol(), coin.ContractID, uint64(0))
	txOutput.Recharge.TxID = tx.TxID
	txOutput.Recharge.Address = to.Name
	txOutput.Recharge.Coin = coin
	txOutput.Recharge.Amount = tx.Amount
	txOutput.Recharge.Symbol = coin.Symbol
	//txOutput.Recharge.IsMemo = true
	//txOutput.Recharge.Memo = data.Memo
	txOutput.Recharge.BlockHash = tx.BlockHash
	txOutput.Recharge.BlockHeight = tx.BlockHeight
	txOutput.Recharge.Index = 0 //账户模型填0
	txOutput.Recharge.CreateAt = time.Now().Unix()
	txExtractData.TxOutputs = append(txExtractData.TxOutputs, txOutput)
}

//newExtractDataNotify 发送通知
func (bs *BtsBlockScanner) newExtractDataNotify(height uint64, extractData map[string][]*openwallet.TxExtractData) error {
	for o := range bs.Observers {
		for key, array := range extractData {
			for _, item := range array {
				err := o.BlockExtractDataNotify(key, item)
				if err != nil {
					log.Error("BlockExtractDataNotify unexpected error:", err)
					//记录未扫区块
					unscanRecord := NewUnscanRecord(height, "", "ExtractData Notify failed.")
					err = bs.SaveUnscanRecord(unscanRecord)
					if err != nil {
						log.Std.Error("block height: %d, save unscan record failed. unexpected error: %v", height, err.Error())
					}

				}
			}

		}
	}

	return nil
}

//ScanBlock 扫描指定高度区块
func (bs *BtsBlockScanner) ScanBlock(height uint64) error {

	block, err := bs.scanBlock(height)
	if err != nil {
		return err
	}

	//通知新区块给观测者，异步处理
	bs.newBlockNotify(block)

	return nil
}

func (bs *BtsBlockScanner) scanBlock(height uint64) (*Block, error) {

	block, err := bs.wm.Api.GetBlockByHeight(uint32(height))
	if err != nil {
		bs.wm.Log.Std.Info("block scanner can not get new block data; unexpected error: %v", err)

		//记录未扫区块
		unscanRecord := NewUnscanRecord(height, "", err.Error())
		bs.SaveUnscanRecord(unscanRecord)
		bs.wm.Log.Std.Info("block height: %d extract failed.", height)
		return nil, err
	}

	bs.wm.Log.Std.Info("block scanner scanning height: %d ...", block.BlockID)

	err = bs.BatchExtractTransactions(block.Height, block.BlockID, block.Timestamp.Unix(), block.Transactions, block.TransactionIDs)
	if err != nil {
		bs.wm.Log.Std.Info("block scanner can not extractRechargeRecords; unexpected error: %v", err)
	}

	return block, nil
}

//SetRescanBlockHeight 重置区块链扫描高度
func (bs *BtsBlockScanner) SetRescanBlockHeight(height uint64) error {
	if height <= 0 {
		return errors.New("block height to rescan must greater than 0. ")
	}

	block, err := bs.wm.Api.GetBlockByHeight(uint32(height - 1))
	if err != nil {
		return err
	}

	bs.SaveLocalBlockHead(uint32(height-1), block.BlockID)

	return nil
}

// GetGlobalMaxBlockHeight GetGlobalMaxBlockHeight
func (bs *BtsBlockScanner) GetGlobalMaxBlockHeight() uint64 {
	headBlock, err := bs.GetGlobalHeadBlock()
	if err != nil {
		bs.wm.Log.Std.Info("get global head block error;unexpected error:%v", err)
		return 0
	}
	return headBlock.Height
}

//GetGlobalHeadBlock GetGlobalHeadBlock
func (bs *BtsBlockScanner) GetGlobalHeadBlock() (block *Block, err error) {
	infoResp, err := bs.GetChainInfo()
	if err != nil {
		bs.wm.Log.Std.Info("get chain info error;unexpected error:%v", err)
		return
	}

	block, err = bs.wm.Api.GetBlockByHeight(uint32(infoResp.HeadBlockNum) - 1)
	if err != nil {
		bs.wm.Log.Std.Info("block scanner can not get block by height; unexpected error:%v", err)
		return
	}

	return
}

//GetChainInfo GetChainInfo
func (bs *BtsBlockScanner) GetChainInfo() (infoResp *BlockchainInfo, err error) {
	infoResp, err = bs.wm.Api.GetBlockchainInfo()
	if err != nil {
		bs.wm.Log.Std.Info("block scanner can not get info; unexpected error:%v", err)
	}
	return
}

//GetScannedBlockHeight 获取已扫区块高度
func (bs *BtsBlockScanner) GetScannedBlockHeight() uint64 {
	height, _, _ := bs.GetLocalBlockHead()
	return uint64(height)
}

//GetBalanceByAddress 查询地址余额
func (bs *BtsBlockScanner) GetBalanceByAddress(address ...string) ([]*openwallet.Balance, error) {

	addrBalanceArr := make([]*openwallet.Balance, 0)

	return addrBalanceArr, nil
}

func (bs *BtsBlockScanner) GetCurrentBlockHeader() (*openwallet.BlockHeader, error) {
	infoResp, err := bs.GetChainInfo()
	if err != nil {
		bs.wm.Log.Std.Info("get chain info error;unexpected error:%v", err)
		return nil, err
	}
	return &openwallet.BlockHeader{Height: uint64(infoResp.HeadBlockNum), Hash: infoResp.HeadBlockID}, nil
}

//rescanFailedRecord 重扫失败记录
func (bs *BtsBlockScanner) RescanFailedRecord() {

	var (
		blockMap = make(map[uint64][]string)
	)

	list, err := bs.wm.GetUnscanRecords()
	if err != nil {
		bs.wm.Log.Std.Info("block scanner can not get rescan data; unexpected error: %v", err)
	}

	//组合成批处理
	for _, r := range list {

		if _, exist := blockMap[r.BlockHeight]; !exist {
			blockMap[r.BlockHeight] = make([]string, 0)
		}

		if len(r.TxID) > 0 {
			arr := blockMap[r.BlockHeight]
			arr = append(arr, r.TxID)

			blockMap[r.BlockHeight] = arr
		}
	}

	for height, _ := range blockMap {

		if height == 0 {
			continue
		}

		bs.wm.Log.Std.Info("block scanner rescanning height: %d ...", height)

		block, err := bs.wm.Api.GetBlockByHeight(uint32(height))
		if err != nil {
			bs.wm.Log.Std.Info("block scanner can not get new block data; unexpected error: %v", err)
			continue
		}

		err = bs.BatchExtractTransactions(height, block.BlockID, block.Timestamp.Unix(), block.Transactions, block.TransactionIDs)
		if err != nil {
			bs.wm.Log.Std.Info("block scanner can not extractRechargeRecords; unexpected error: %v", err)
			continue
		}

		//删除未扫记录
		bs.DeleteUnscanRecord(uint32(height))
	}

	//删除未没有找到交易记录的重扫记录
	bs.wm.DeleteUnscanRecordNotFindTX()
}
