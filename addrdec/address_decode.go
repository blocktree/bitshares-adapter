package addrdec

import (
	"fmt"
	"github.com/blocktree/openwallet/openwallet"
	"strings"

	"github.com/blocktree/go-owcdrivers/addressEncoder"
)

var (
	BTSPublicKeyPrefix       = "PUB_"
	BTSPublicKeyK1Prefix     = "PUB_K1_"
	BTSPublicKeyR1Prefix     = "PUB_R1_"
	BTSPublicKeyPrefixCompat = "BTS"

	//BTS stuff
	BTS_mainnetPublic = addressEncoder.AddressType{"bts", addressEncoder.BTCAlphabet, "ripemd160", "", 33, []byte(BTSPublicKeyPrefixCompat), nil}
	// BTS_mainnetPrivateWIF           = AddressType{"base58", BTCAlphabet, "doubleSHA256", "", 32, []byte{0x80}, nil}
	// BTS_mainnetPrivateWIFCompressed = AddressType{"base58", BTCAlphabet, "doubleSHA256", "", 32, []byte{0x80}, []byte{0x01}}

	Default = AddressDecoderV2{}
)

//AddressDecoderV2
type AddressDecoderV2 struct {
	*openwallet.AddressDecoderV2Base
	IsTestNet bool
}

//NewAddressDecoder 地址解析器
func NewAddressDecoderV2() *AddressDecoderV2 {
	decoder := AddressDecoderV2{}
	return &decoder
}

// AddressDecode decode address
func (dec *AddressDecoderV2) AddressDecode(pubKey string, opts ...interface{}) ([]byte, error) {

	var pubKeyMaterial string
	if strings.HasPrefix(pubKey, BTSPublicKeyR1Prefix) {
		pubKeyMaterial = pubKey[len(BTSPublicKeyR1Prefix):] // strip "PUB_R1_"
	} else if strings.HasPrefix(pubKey, BTSPublicKeyK1Prefix) {
		pubKeyMaterial = pubKey[len(BTSPublicKeyK1Prefix):] // strip "PUB_K1_"
	} else if strings.HasPrefix(pubKey, BTSPublicKeyPrefixCompat) { // "BTS"
		pubKeyMaterial = pubKey[len(BTSPublicKeyPrefixCompat):] // strip "BTS"
	} else {
		return nil, fmt.Errorf("public key should start with [%q | %q] (or the old %q)", BTSPublicKeyK1Prefix, BTSPublicKeyR1Prefix, BTSPublicKeyPrefixCompat)
	}

	ret, err := addressEncoder.Base58Decode(pubKeyMaterial, addressEncoder.NewBase58Alphabet(BTS_mainnetPublic.Alphabet))
	if err != nil {
		return nil, addressEncoder.ErrorInvalidAddress
	}
	if addressEncoder.VerifyChecksum(ret, BTS_mainnetPublic.ChecksumType) == false {
		return nil, addressEncoder.ErrorInvalidAddress
	}

	return ret[:len(ret)-4], nil
}

// AddressEncode encode address
func (dec *AddressDecoderV2) AddressEncode(hash []byte, opts ...interface{}) (string, error) {
	data := addressEncoder.CatData(hash, addressEncoder.CalcChecksum(hash, BTS_mainnetPublic.ChecksumType))
	return string(BTS_mainnetPublic.Prefix) + addressEncoder.EncodeData(data, "base58", BTS_mainnetPublic.Alphabet), nil
}
