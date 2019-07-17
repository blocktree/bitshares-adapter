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

var (
	tw *WalletManager
)

func init() {
	tw = testNewWalletManager()
}

func testNewWalletManager() *WalletManager {
	wm := NewWalletManager(nil)
	wm.Config.ServerAPI = "http://127.0.0.1:8080/rpc"
	wm.Api = NewWalletClient(wm.Config.ServerAPI, "", false)
	return wm
}

// func TestWalletManager_GetInfo(t *testing.T) {
// 	wm := testNewWalletManager()
// 	r, err := wm.Api.GetInfo()
// 	if err != nil {
// 		log.Errorf("unexpected error: %v", err)
// 		return
// 	}
// 	log.Infof("%+v", r)
// }

// func TestWalletManager_GetAccount(t *testing.T) {
// 	wm := testNewWalletManager()
// 	r, err := wm.Api.GetAccount("eostesterkkk")
// 	if err != nil {
// 		log.Errorf("unexpected error: %v", err)
// 		return
// 	}
// 	log.Infof("%+v", r)
// }

// func TestWalletManager_GetBlock(t *testing.T) {
// 	wm := testNewWalletManager()
// 	r, err := wm.Api.GetBlockByNum(1000000)
// 	if err != nil {
// 		log.Errorf("unexpected error: %v", err)
// 		return
// 	}
// 	log.Infof("%+v", r)
// }

// func TestWalletManager_GetTransaction(t *testing.T) {
// 	wm := testNewWalletManager()
// 	r, err := wm.Api.GetTransaction("498db83f0358ff7a9e32215d924be4df3e84bf7fef1614e241fb764afe3ecd8d")
// 	if err != nil {
// 		log.Errorf("unexpected error: %v", err)
// 		return
// 	}
// 	log.Infof("%+v", r)
// }

// func TestWalletManager_GetABI(t *testing.T) {
// 	wm := testNewWalletManager()
// 	r, err := wm.Api.GetABI("bitshares.token")
// 	if err != nil {
// 		log.Errorf("unexpected error: %v", err)
// 		return
// 	}
// 	log.Infof("%+v", r)
// }

// func TestWalletManager_GetCurrencyBalance(t *testing.T) {
// 	wm := testNewWalletManager()
// 	r, err := wm.Api.GetCurrencyBalance("bob", "EOS", "bitshares.token")
// 	if err != nil {
// 		log.Errorf("unexpected error: %v", err)
// 		return
// 	}
// 	log.Infof("%+v", r)
// }
