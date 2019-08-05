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
	"encoding/hex"
	"testing"

	"github.com/blocktree/bitshares-adapter/addrdec"
)

func TestAddressDecoder_AddressEncode(t *testing.T) {
	addrdec.Default.IsTestNet = false

	p2pk, _ := hex.DecodeString("5b1ac00a18dc9bb1a0494206d9ad72ef7fed0873")
	p2pkAddr, _ := addrdec.Default.AddressEncode(p2pk)
	t.Logf("p2pkAddr: %s", p2pkAddr)

	p2sh, _ := hex.DecodeString("adf2c47cbf6a053ebf45b033ba2044c36984a468")
	p2shAddr, _ := addrdec.Default.AddressEncode(p2sh)
	t.Logf("p2shAddr: %s", p2shAddr)
}

func TestAddressDecoder_AddressDecode(t *testing.T) {

	addrdec.Default.IsTestNet = false

	p2pkAddr := "D8qHfnugKAgavULzVjQKyjqxMD7wBETN4s"
	p2pkHash, _ := addrdec.Default.AddressDecode(p2pkAddr)
	t.Logf("p2pkHash: %s", hex.EncodeToString(p2pkHash))

	p2shAddr := "Lb5hzBamSSS4xz2FJqAV2cL2bYMq6oDjJA"

	p2shHash, _ := addrdec.Default.AddressDecode(p2shAddr)
	t.Logf("p2shHash: %s", hex.EncodeToString(p2shHash))
}
