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
	"testing"

	"github.com/blocktree/openwallet/v2/openwallet"
)

func TestBtsBlockScanner_ScanBlock(t *testing.T) {
	type fields struct {
		BlockScannerBase     *openwallet.BlockScannerBase
		CurrentBlockHeight   uint64
		extractingCH         chan struct{}
		wm                   *WalletManager
		IsScanMemPool        bool
		RescanLastBlockCount uint64
	}
	type args struct {
		height uint64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bs := &BtsBlockScanner{
				BlockScannerBase:     tt.fields.BlockScannerBase,
				CurrentBlockHeight:   tt.fields.CurrentBlockHeight,
				extractingCH:         tt.fields.extractingCH,
				wm:                   tt.fields.wm,
				IsScanMemPool:        tt.fields.IsScanMemPool,
				RescanLastBlockCount: tt.fields.RescanLastBlockCount,
			}
			if err := bs.ScanBlock(tt.args.height); (err != nil) != tt.wantErr {
				t.Errorf("BtsBlockScanner.ScanBlock() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
