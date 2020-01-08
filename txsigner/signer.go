package txsigner

var Default = &TransactionSigner{}

type TransactionSigner struct {
}

// SignTransactionHash 交易哈希签名算法
// Deprecated
func (singer *TransactionSigner) SignTransactionHash(msg []byte, privateKey []byte, eccType uint32) ([]byte, error) {
	//return SignCanonical(privateKey, msg)
	return nil, nil
}

// Deprecated
func (singer *TransactionSigner) VerifyAndCombineSignature(msg, publicKey, signature []byte) (bool, []byte, error) {
	//compactSig, err := makeCompact(signature, publicKey, msg)
	//if err != nil {
	//	return false, nil, err
	//}
	//
	//if !isCanonical(compactSig) {
	//	return false, nil, fmt.Errorf("it is not canonical signature")
	//}
	//
	//return true, compactSig, nil
	return false, nil, nil
}
