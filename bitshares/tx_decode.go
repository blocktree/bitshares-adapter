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
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/blocktree/bitshares-adapter/encoding"
	"github.com/blocktree/bitshares-adapter/types"
	"github.com/denkhaus/bitshares/config"
	"github.com/denkhaus/bitshares/crypto"
	"github.com/denkhaus/bitshares/operations"
	bt "github.com/denkhaus/bitshares/types"

	owcrypt "github.com/blocktree/go-owcrypt"
	"github.com/blocktree/openwallet/v2/openwallet"
	"github.com/shopspring/decimal"
)

// TransactionDecoder 交易单解析器
type TransactionDecoder struct {
	openwallet.TransactionDecoderBase
	wm *WalletManager //钱包管理者
}

//NewTransactionDecoder 交易单解析器
func NewTransactionDecoder(wm *WalletManager) *TransactionDecoder {
	decoder := TransactionDecoder{}
	decoder.wm = wm
	return &decoder
}

//CreateRawTransaction 创建交易单
func (decoder *TransactionDecoder) CreateRawTransaction(wrapper openwallet.WalletDAI, rawTx *openwallet.RawTransaction) error {

	var (
		accountID = rawTx.Account.AccountID
		amountStr string
		to        string
		assetID   types.ObjectID
		precise   uint64
	)

	assetID = types.MustParseObjectID(rawTx.Coin.Contract.Address)
	precise = rawTx.Coin.Contract.Decimals

	//获取wallet
	account, err := wrapper.GetAssetsAccountInfo(accountID)
	if err != nil {
		return err
	}

	if account.Alias == "" {
		return fmt.Errorf("[%s] have not been created", accountID)
	}

	for k, v := range rawTx.To {
		amountStr = v
		to = k
		break
	}

	// 检查转出、目标账户是否存在
	accounts, err := decoder.wm.Api.GetAccounts(account.Alias, to)
	if err != nil {
		return openwallet.Errorf(openwallet.ErrAccountNotAddress, "accounts have not registered [%v] ", err)
	}

	fromAccount := accounts[0]
	toAccount := accounts[1]

	// 检查转出账户余额
	balance, err := decoder.wm.Api.GetAssetsBalance(fromAccount.ID, assetID)
	if err != nil || balance == nil {
		return openwallet.Errorf(openwallet.ErrInsufficientBalanceOfAccount, "all address's balance of account is not enough")
	}

	accountBalanceDec, _ := decimal.NewFromString(balance.Amount)
	amountDec, _ := decimal.NewFromString(amountStr)
	amountDec = amountDec.Shift(int32(precise))

	if accountBalanceDec.LessThan(amountDec) {
		return fmt.Errorf("the balance: %s is not enough", amountStr)
	}

	memo := rawTx.GetExtParam().Get("memo").String()

	asset := bt.AssetIDFromObject(bt.NewAssetID(assetID.String()))
	amount := bt.AssetAmount{
		Asset:  asset,
		Amount: bt.Int64(amountDec.IntPart()),
	}

	op := operations.TransferOperation{
		Amount:     amount,
		Extensions: bt.Extensions{},
		From:       bt.AccountIDFromObject(bt.NewAccountID(fromAccount.ID.String())),
		To:         bt.AccountIDFromObject(bt.NewAccountID(toAccount.ID.String())),
	}

	fromPublicKey, _ := bt.NewPublicKeyFromString(fromAccount.Options.MemoKey)
	toPublicKey, _ := bt.NewPublicKeyFromString(toAccount.Options.MemoKey)

	if memo != "" {
		m := bt.Memo{
			From:  *fromPublicKey,
			To:    *toPublicKey,
			Nonce: bt.UInt64(rand.Uint64()),
		}
		wif := decoder.wm.Config.MemoPrivateKey
		encoding.Encrypt(&m, memo, wif)
		decoder.wm.Log.Debug("memo hash:", m.Message)
		op.Memo = &m
	}

	ops := &bt.Operations{&op}

	createTxErr := decoder.createRawTransaction(
		wrapper,
		rawTx,
		&accountBalanceDec,
		account.Alias,
		ops,
		memo)
	if createTxErr != nil {
		return createTxErr
	}

	return nil

}

//SignRawTransaction 签名交易单
func (decoder *TransactionDecoder) SignRawTransaction(wrapper openwallet.WalletDAI, rawTx *openwallet.RawTransaction) error {

	if rawTx.Signatures == nil || len(rawTx.Signatures) == 0 {
		//this.wm.Log.Std.Error("len of signatures error. ")
		return fmt.Errorf("transaction signature is empty")
	}

	key, err := wrapper.HDKey()
	if err != nil {
		return err
	}

	keySignatures := rawTx.Signatures[rawTx.Account.AccountID]
	if keySignatures != nil {
		for _, keySignature := range keySignatures {

			childKey, err := key.DerivedKeyWithPath(keySignature.Address.HDPath, keySignature.EccType)
			keyBytes, err := childKey.GetPrivateKeyBytes()
			if err != nil {
				return err
			}

			hash, err := hex.DecodeString(keySignature.Message)
			if err != nil {
				return fmt.Errorf("decoder transaction hash failed, unexpected err: %v", err)
			}

			decoder.wm.Log.Debug("hash:", hash)

			//sig, err := txsigner.Default.SignTransactionHash(hash, keyBytes, decoder.wm.CurveType())
			//if err != nil {
			//	return fmt.Errorf("sign transaction hash failed, unexpected err: %v", err)
			//}
			signature, v, sigErr := owcrypt.Signature(keyBytes, nil, hash, decoder.wm.CurveType())
			if sigErr == owcrypt.FAILURE {
				return fmt.Errorf("sign transaction hash failed")
			}
			signature = append(signature, v)

			keySignature.Signature = hex.EncodeToString(signature)
		}
	}

	decoder.wm.Log.Info("transaction hash sign success")

	rawTx.Signatures[rawTx.Account.AccountID] = keySignatures

	return nil
}

//VerifyRawTransaction 验证交易单，验证交易单并返回加入签名后的交易单
func (decoder *TransactionDecoder) VerifyRawTransaction(wrapper openwallet.WalletDAI, rawTx *openwallet.RawTransaction) error {

	if rawTx.Signatures == nil || len(rawTx.Signatures) == 0 {
		//this.wm.Log.Std.Error("len of signatures error. ")
		return fmt.Errorf("transaction signature is empty")
	}

	var tx bt.SignedTransaction
	txHex, err := hex.DecodeString(rawTx.RawHex)
	if err != nil {
		return fmt.Errorf("transaction DecodeString failed, unexpected error: %v", err)
	}
	err = tx.UnmarshalJSON(txHex)
	if err != nil {
		return fmt.Errorf("transaction UnmarshalJSON failed, unexpected error: %v", err)
	}

	//支持多重签名
	for accountID, keySignatures := range rawTx.Signatures {
		decoder.wm.Log.Debug("accountID Signatures:", accountID)
		for _, keySignature := range keySignatures {

			messsage, _ := hex.DecodeString(keySignature.Message)
			signature, _ := hex.DecodeString(keySignature.Signature)
			//publicKey, _ := hex.DecodeString(keySignature.Address.PublicKey)

			//验证签名，解压公钥，解压后首字节04要去掉
			//uncompessedPublicKey := owcrypt.PointDecompress(publicKey, decoder.wm.CurveType())

			//valid, compactSig, err := txsigner.Default.VerifyAndCombineSignature(messsage, uncompessedPublicKey[1:], signature)
			//if !valid {
			//	return fmt.Errorf("transaction verify failed: %v", err)
			//}

			_, valid := owcrypt.RecoverPubkey(signature, messsage, decoder.wm.CurveType())
			//valid, compactSig, err := eos_txsigner.Default.VerifyAndCombineSignature(messsage, uncompessedPublicKey[1:], signature)
			if valid == owcrypt.FAILURE {
				return fmt.Errorf("transaction verify failed: %v", err)
			}
			v := signature[len(signature)-1] //签名最后一字节是v

			//验签通过后处理V值，符合节点验签
			compactSig := signature[:len(signature)-1]
			compactSig = append([]byte{v + 27 + 4}, compactSig...)

			tx.Signatures = append(
				tx.Signatures,
				compactSig,
			)
		}
	}

	rawTx.IsCompleted = true
	jsonTx, _ := tx.MarshalJSON()
	rawTx.RawHex = hex.EncodeToString(jsonTx)

	return nil
}

// SubmitRawTransaction 广播交易单
func (decoder *TransactionDecoder) SubmitRawTransaction(wrapper openwallet.WalletDAI, rawTx *openwallet.RawTransaction) (*openwallet.Transaction, error) {

	var stx bt.SignedTransaction
	txHex, err := hex.DecodeString(rawTx.RawHex)
	if err != nil {
		return nil, fmt.Errorf("transaction decode hex failed, unexpected error: %v", err)
	}
	err = stx.UnmarshalJSON(txHex)
	if err != nil {
		return nil, fmt.Errorf("transaction decode json failed, unexpected error: %v", err)
	}

	resp, err := decoder.wm.Api.BroadcastTransaction(&stx)
	if err != nil {
		return nil, fmt.Errorf("push transaction: %s", err)
	}

	decoder.wm.Log.Info("Transaction [%s] submitted to the network successfully.", resp.ID)

	rawTx.TxID = resp.ID
	rawTx.IsSubmit = true

	decimals := int32(rawTx.Coin.Contract.Decimals)
	fees := "0"

	//记录一个交易单
	tx := &openwallet.Transaction{
		From:       rawTx.TxFrom,
		To:         rawTx.TxTo,
		Amount:     rawTx.TxAmount,
		Coin:       rawTx.Coin,
		TxID:       rawTx.TxID,
		Decimal:    decimals,
		AccountID:  rawTx.Account.AccountID,
		Fees:       fees,
		SubmitTime: time.Now().Unix(),
		ExtParam:   rawTx.ExtParam,
	}

	tx.WxID = openwallet.GenTransactionWxID(tx)

	return tx, nil
}

//GetRawTransactionFeeRate 获取交易单的费率
func (decoder *TransactionDecoder) GetRawTransactionFeeRate() (feeRate string, unit string, err error) {
	return "", "", nil
}

//CreateSummaryRawTransaction 创建汇总交易
func (decoder *TransactionDecoder) CreateSummaryRawTransaction(wrapper openwallet.WalletDAI, sumRawTx *openwallet.SummaryRawTransaction) ([]*openwallet.RawTransaction, error) {
	var (
		rawTxWithErrArray []*openwallet.RawTransactionWithError
		rawTxArray        = make([]*openwallet.RawTransaction, 0)
		err               error
	)
	rawTxWithErrArray, err = decoder.CreateSummaryRawTransactionWithError(wrapper, sumRawTx)
	if err != nil {
		return nil, err
	}
	for _, rawTxWithErr := range rawTxWithErrArray {
		if rawTxWithErr.Error != nil {
			continue
		}
		rawTxArray = append(rawTxArray, rawTxWithErr.RawTx)
	}
	return rawTxArray, nil
}

//CreateSummaryRawTransactionWithError 创建汇总交易
func (decoder *TransactionDecoder) CreateSummaryRawTransactionWithError(wrapper openwallet.WalletDAI, sumRawTx *openwallet.SummaryRawTransaction) ([]*openwallet.RawTransactionWithError, error) {

	var (
		rawTxArray = make([]*openwallet.RawTransactionWithError, 0)
		accountID  = sumRawTx.Account.AccountID
		assetID    types.ObjectID
		precise    uint64
	)

	assetID = types.MustParseObjectID(sumRawTx.Coin.Contract.Address)
	precise = sumRawTx.Coin.Contract.Decimals

	minTransfer, _ := decimal.NewFromString(sumRawTx.MinTransfer)
	retainedBalance, _ := decimal.NewFromString(sumRawTx.RetainedBalance)

	if minTransfer.LessThan(retainedBalance) {
		return nil, fmt.Errorf("mini transfer amount must be greater than address retained balance")
	}

	//获取wallet
	account, err := wrapper.GetAssetsAccountInfo(accountID)
	if err != nil {
		return nil, err
	}

	if account.Alias == "" {
		return nil, fmt.Errorf("[%s] have not been created", accountID)
	}

	// 检查转出、目标账户是否存在
	accounts, err := decoder.wm.Api.GetAccounts(account.Alias, sumRawTx.SummaryAddress)
	if err != nil {
		return nil, openwallet.Errorf(openwallet.ErrAccountNotAddress, "accounts have not registered [%v] ", err)
	}

	fromAccount := accounts[0]
	toAccount := accounts[1]

	// 检查转出账户余额
	balance, err := decoder.wm.Api.GetAssetsBalance(fromAccount.ID, assetID)
	if err != nil || balance == nil {
		return nil, openwallet.Errorf(openwallet.ErrInsufficientBalanceOfAccount, "all address's balance of account is not enough")
	}

	accountBalanceDec, _ := decimal.NewFromString(balance.Amount)
	minTransfer = minTransfer.Shift(int32(precise))
	retainedBalance = retainedBalance.Shift(int32(precise))

	if accountBalanceDec.LessThan(minTransfer) || accountBalanceDec.LessThanOrEqual(decimal.Zero) {
		return rawTxArray, nil
	}

	//计算汇总数量 = 余额 - 保留余额
	sumAmount := accountBalanceDec.Sub(retainedBalance)

	amountInt64 := sumAmount.IntPart()
	memo := sumRawTx.GetExtParam().Get("memo").String()

	decoder.wm.Log.Debugf("balance: %v", accountBalanceDec.String())
	decoder.wm.Log.Debugf("fees: %d", 0)
	decoder.wm.Log.Debugf("sumAmount: %v", sumAmount)

	//创建一笔交易单
	rawTx := &openwallet.RawTransaction{
		Coin:    sumRawTx.Coin,
		Account: sumRawTx.Account,
		To: map[string]string{
			sumRawTx.SummaryAddress: sumAmount.String(),
		},
		Required: 1,
	}

	asset := bt.AssetIDFromObject(bt.NewAssetID(assetID.String()))
	amount := bt.AssetAmount{
		Asset:  asset,
		Amount: bt.Int64(amountInt64),
	}

	op := operations.TransferOperation{
		Amount:     amount,
		Extensions: bt.Extensions{},
		From:       bt.AccountIDFromObject(bt.NewAccountID(fromAccount.ID.String())),
		To:         bt.AccountIDFromObject(bt.NewAccountID(toAccount.ID.String())),
	}

	fromPublicKey, _ := bt.NewPublicKeyFromString(fromAccount.Options.MemoKey)
	toPublicKey, _ := bt.NewPublicKeyFromString(toAccount.Options.MemoKey)

	if memo != "" {
		m := bt.Memo{
			From:  *fromPublicKey,
			To:    *toPublicKey,
			Nonce: bt.UInt64(rand.Uint64()),
		}
		keyBag := crypto.NewKeyBag()
		keyBag.Add(decoder.wm.Config.MemoPrivateKey)

		if err := keyBag.EncryptMemo(&m, memo); err != nil {
			return nil, fmt.Errorf("EncryptMemo: %v", err)
		}

		op.Memo = &m
	}

	ops := &bt.Operations{&op}

	createTxErr := decoder.createRawTransaction(
		wrapper,
		rawTx,
		&accountBalanceDec,
		account.Alias,
		ops,
		memo)
	rawTxWithErr := &openwallet.RawTransactionWithError{
		RawTx: rawTx,
		Error: createTxErr,
	}

	//创建成功，添加到队列
	rawTxArray = append(rawTxArray, rawTxWithErr)

	return rawTxArray, nil
}

//createRawTransaction
func (decoder *TransactionDecoder) createRawTransaction(
	wrapper openwallet.WalletDAI,
	rawTx *openwallet.RawTransaction,
	balanceDec *decimal.Decimal,
	from string,
	ops *bt.Operations,
	memo string) *openwallet.Error {

	var (
		to               string
		accountTotalSent = decimal.Zero
		txFrom           = make([]string, 0)
		txTo             = make([]string, 0)
		keySignList      = make([]*openwallet.KeySignature, 0)
		accountID        = rawTx.Account.AccountID
		amountDec        = decimal.Zero
		curveType        = decoder.wm.Config.CurveType
		assetID          = bt.NewAssetID(rawTx.Coin.Contract.Address)
		precise          = rawTx.Coin.Contract.Decimals
		operations       = bt.Operations(*ops)
	)

	for k, v := range rawTx.To {
		to = k
		amountDec, _ = decimal.NewFromString(v)
		break
	}

	fees, err := decoder.wm.Api.GetRequiredFee(operations, assetID.String())
	if err != nil {
		return openwallet.Errorf(openwallet.ErrCreateRawTransactionFailed, "can't get fees")
	}

	feesDec := decimal.Zero
	for _, fee := range fees {
		feesDec = feesDec.Add(decimal.New(int64(fee.Amount), 0))
	}

	if balanceDec.LessThan(amountDec.Add(feesDec)) {
		return openwallet.Errorf(openwallet.ErrCreateRawTransactionFailed, "the balance: %s is not enough", balanceDec.Shift(-int32(precise)))
	}

	if err := operations.ApplyFees(fees); err != nil {
		return openwallet.Errorf(openwallet.ErrCreateRawTransactionFailed, "ApplyFees")
	}

	info, err := decoder.wm.Api.GetBlockchainInfo()
	if err != nil {
		return openwallet.Errorf(openwallet.ErrCreateRawTransactionFailed, "GetBlockchainInfo: %v", err)
	}

	j, _ := json.Marshal(info.HeadBlockID)
	headBlockID := bt.String{}
	headBlockID.UnmarshalJSON(j)
	props := &bt.DynamicGlobalProperties{
		HeadBlockID:              headBlockID,
		HeadBlockNumber:          bt.UInt32(info.HeadBlockNum),
		LastIrreversibleBlockNum: bt.UInt32(info.LastIrreversibleBlockNum),
	}

	tx, err := bt.NewSignedTransactionWithBlockData(props)
	if err != nil {
		return openwallet.Errorf(openwallet.ErrCreateRawTransactionFailed, "NewTransaction")
	}

	tx.Operations = operations
	signer := crypto.NewTransactionSigner(tx)

	//交易哈希
	digest, err := signer.Digest(config.Current())
	if err != nil {
		return openwallet.Errorf(openwallet.ErrCreateRawTransactionFailed, "Calculate digest error: %v", err)
	}

	addresses, err := wrapper.GetAddressList(0, -1,
		"AccountID", accountID)
	if err != nil {
		return openwallet.ConvertError(err)
	}

	if len(addresses) == 0 {
		return openwallet.Errorf(openwallet.ErrCreateRawTransactionFailed, "[%s] have not public key", accountID)
	}

	for _, addr := range addresses {
		signature := openwallet.KeySignature{
			EccType: curveType,
			Nonce:   "",
			Address: addr,
			Message: hex.EncodeToString(digest),
			RSV:     true,
		}
		keySignList = append(keySignList, &signature)
	}

	//计算账户的实际转账amount
	if from != to {
		accountTotalSent = accountTotalSent.Add(amountDec)
	}
	accountTotalSent = decimal.Zero.Sub(accountTotalSent)

	txFrom = []string{fmt.Sprintf("%s:%s", from, amountDec.String())}
	txTo = []string{fmt.Sprintf("%s:%s", to, amountDec.String())}

	if rawTx.Signatures == nil {
		rawTx.Signatures = make(map[string][]*openwallet.KeySignature)
	}

	jsonTx, _ := tx.MarshalJSON()
	rawTx.RawHex = hex.EncodeToString(jsonTx)
	rawTx.Signatures[rawTx.Account.AccountID] = keySignList
	rawTx.FeeRate = "0"
	rawTx.Fees = feesDec.String()
	rawTx.IsBuilt = true
	rawTx.TxAmount = accountTotalSent.String()
	rawTx.TxFrom = txFrom
	rawTx.TxTo = txTo

	return nil
}
