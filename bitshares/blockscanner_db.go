/*
 * Copyright 2018 The openwallet Authors
 * This file is part of the openwallet library.
 *
 * The openwallet library is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The openwallet library is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU Lesser General Public License for more details.
 */

package bitshares

import (
	"errors"
	"path/filepath"

	"github.com/asdine/storm"
)

//SaveLocalBlockHead 记录区块高度和hash到本地
func (bs *BtsBlockScanner) SaveLocalBlockHead(blockHeight uint32, blockHash string) error {

	//获取本地区块高度
	db, err := storm.Open(filepath.Join(bs.wm.Config.dbPath, bs.wm.Config.BlockchainFile))
	if err != nil {
		return err
	}
	defer db.Close()

	db.Set(blockchainBucket, "blockHeight", &blockHeight)
	db.Set(blockchainBucket, "blockHash", &blockHash)

	return nil
}

//GetLocalBlockHead 获取本地记录的区块高度和hash
func (bs *BtsBlockScanner) GetLocalBlockHead() (uint32, string, error) {

	var (
		blockHeight uint32
		blockHash   string
	)

	//获取本地区块高度
	db, err := storm.Open(filepath.Join(bs.wm.Config.dbPath, bs.wm.Config.BlockchainFile))
	if err != nil {
		return 0, "", err
	}
	defer db.Close()

	db.Get(blockchainBucket, "blockHeight", &blockHeight)
	db.Get(blockchainBucket, "blockHash", &blockHash)

	return blockHeight, blockHash, nil
}

//SaveLocalBlock 记录本地新区块
func (bs *BtsBlockScanner) SaveLocalBlock(blockHeader *Block) error {

	db, err := storm.Open(filepath.Join(bs.wm.Config.dbPath, bs.wm.Config.BlockchainFile))
	if err != nil {
		return err
	}
	defer db.Close()

	db.Save(blockHeader)

	return nil
}

//GetLocalBlock 获取本地区块数据
func (bs *BtsBlockScanner) GetLocalBlock(height uint32) (*Block, error) {

	var (
		blockHeader Block
	)

	db, err := storm.Open(filepath.Join(bs.wm.Config.dbPath, bs.wm.Config.BlockchainFile))
	if err != nil {
		return nil, err
	}
	defer db.Close()

	err = db.One("Height", height, &blockHeader)
	if err != nil {
		return nil, err
	}

	return &blockHeader, nil
}

//SaveUnscanRecord 保存交易记录到钱包数据库
func (bs *BtsBlockScanner) SaveUnscanRecord(record *UnscanRecord) error {

	if record == nil {
		return errors.New("the unscan record to save is nil")
	}

	//获取本地区块高度
	db, err := storm.Open(filepath.Join(bs.wm.Config.dbPath, bs.wm.Config.BlockchainFile))
	if err != nil {
		return err
	}
	defer db.Close()

	return db.Save(record)
}

//DeleteUnscanRecord 删除指定高度的未扫记录
func (bs *BtsBlockScanner) DeleteUnscanRecord(height uint32) error {
	//获取本地区块高度
	db, err := storm.Open(filepath.Join(bs.wm.Config.dbPath, bs.wm.Config.BlockchainFile))
	if err != nil {
		return err
	}
	defer db.Close()

	var list []*UnscanRecord
	err = db.Find("BlockHeight", height, &list)
	if err != nil {
		return err
	}

	for _, r := range list {
		db.DeleteStruct(r)
	}

	return nil
}
