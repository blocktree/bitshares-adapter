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
	"fmt"
	"time"

	"github.com/blocktree/bitshares-adapter/types"
	"github.com/blocktree/openwallet/common"
	"github.com/blocktree/openwallet/crypto"
	"github.com/tidwall/gjson"
)

type Asset struct {
	ID                 types.ObjectID `json:"id"`
	Symbol             string         `json:"symbol"`
	Precision          uint8          `json:"precision"`
	Issuer             string         `json:"issuer"`
	DynamicAssetDataID string         `json:"dynamic_asset_data_id"`
}

type BlockHeader struct {
	TransactionMerkleRoot string            `json:"transaction_merkle_root"`
	Previous              string            `json:"previous"`
	Timestamp             types.Time        `json:"timestamp"`
	Witness               string            `json:"witness"`
	Extensions            []json.RawMessage `json:"extensions"`
}

type Block struct {
	Height                uint64
	BlockID               string              `json:"block_id"`
	TransactionMerkleRoot string              `json:"transaction_merkle_root"`
	Previous              string              `json:"previous"`
	Timestamp             types.Time          `json:"timestamp"`
	Witness               string              `json:"witness"`
	Extensions            []json.RawMessage   `json:"extensions"`
	WitnessSignature      string              `json:"witness_signature"`
	Transactions          []types.Transaction `json:"transactions"`
}

func NewBlock(height uint64, result *gjson.Result) *Block {
	obj := Block{}
	json.Unmarshal([]byte(result.Raw), &obj)
	return &obj
}

func NewTransaction(result *gjson.Result) (*types.Transaction, error) {
	obj := types.Transaction{}
	err := json.Unmarshal([]byte(result.Raw), &obj)
	return &obj, err
}

//UnscanRecord 扫描失败的区块及交易
type UnscanRecord struct {
	ID          string `storm:"id"` // primary key
	BlockHeight uint64
	TxID        string
	Reason      string
}

//NewUnscanRecord new UnscanRecord
func NewUnscanRecord(height uint64, txID, reason string) *UnscanRecord {
	obj := UnscanRecord{}
	obj.BlockHeight = height
	obj.TxID = txID
	obj.Reason = reason
	obj.ID = common.Bytes2Hex(crypto.SHA256([]byte(fmt.Sprintf("%d_%s", height, txID))))
	return &obj
}

// TransferAction transfer action
type TransferAction struct {
	TransferData
}

// TransferData token contract transfer action data
type TransferData struct {
	From string `json:"from,omitempty"`
	To   string `json:"to,omitempty"`
	Memo string `json:"memo,omitempty"`
}

// // ParseHeader 区块链头
// func ParseHeader(b *eos.BlockResp) *openwallet.BlockHeader {
// 	obj := openwallet.BlockHeader{}

// 	//解析josn
// 	obj.Merkleroot = b.TransactionMRoot.String()
// 	obj.Hash = b.ID.String()
// 	obj.Previousblockhash = b.Previous.String()
// 	obj.Height = uint64(b.BlockNum)
// 	obj.Time = uint64(b.Timestamp.Unix())
// 	obj.Symbol = Symbol
// 	return &obj
// }

// // ParseBlock 区块
// func ParseBlock(b *eos.BlockResp) *Block {
// 	obj := Block{}

// 	//解析josn
// 	obj.Merkleroot = b.TransactionMRoot.String()
// 	obj.Hash = b.ID.String()
// 	obj.Previousblockhash = b.Previous.String()
// 	obj.Height = uint32(b.BlockNum)
// 	obj.Time = uint64(b.Timestamp.Unix())
// 	obj.Symbol = Symbol
// 	return &obj
// }

type BlockchainInfo struct {
	HeadBlockNum uint64    `json:"head_block_number"`
	HeadBlockID  string    `json:"head_block_id"`
	Timestamp    time.Time `json:"time"`

	/*
		{
			"id": "2.1.0",
			"head_block_number": 1544081,
			"head_block_id": "00178f912d70e9ed3539f2acfba4752dee5d77bb",
			"time": "2019-07-17T04:09:40",
			"current_witness": "1.6.8",
			"next_maintenance_time": "2019-07-18T00:00:00",
			"last_budget_time": "2019-07-17T00:00:00",
			"witness_budget": 0,
			"accounts_registered_this_interval": 2,
			"recently_missed_count": 0,
			"current_aslot": 1672768,
			"recent_slots_filled": "340282366920938463463374607431768211455",
			"dynamic_flags": 0,
			"last_irreversible_block_num": 1544074
		}
	*/
}

const TimeLayout = `2006-01-02T15:04:05`

func NewBlockchainInfo(result *gjson.Result) *BlockchainInfo {
	obj := BlockchainInfo{}
	arr := result.Array()
	if len(arr) > 0 {
		obj.HeadBlockNum = arr[0].Get("head_block_number").Uint()
		obj.HeadBlockID = arr[0].Get("head_block_id").String()
		obj.Timestamp, _ = time.ParseInLocation(TimeLayout, arr[0].Get("time").String(), time.UTC)
	}
	return &obj
}
