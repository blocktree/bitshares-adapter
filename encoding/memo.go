package encoding

import (
	"fmt"

	"github.com/blocktree/bitshares-adapter/exception"
	"github.com/denkhaus/bitshares/crypto"
	"github.com/denkhaus/bitshares/types"
)

// Encrypt message of a operation.
func Encrypt(memo *types.Memo, msg string, wif string) error {
	keyBag := crypto.NewKeyBag()
	keyBag.Add(wif)

	if err := keyBag.EncryptMemo(memo, msg); err != nil {
		return fmt.Errorf("EncryptMemo: %v", err)
	}
	return nil
}

//Decrypt calculates a shared secret by the receivers private key
//and the senders public key, then decrypts the given memo message.
func Decrypt(msg []byte, fromPub, toPub string, nonce uint64, wif string) (string, error) {

	if len(msg) == 0 || len(fromPub) == 0 || len(toPub) == 0 {
		return "", fmt.Errorf("args is empty")
	}

	if len(wif) == 0 {
		return "", fmt.Errorf("wif cannot be empty")
	}

	from, err := types.NewPublicKeyFromString(fromPub)
	if err != nil {
		return "", fmt.Errorf("NewPublicKeyFromString: %v", err)
	}
	to, err := types.NewPublicKeyFromString(toPub)
	if err != nil {
		return "", fmt.Errorf("NewPublicKeyFromString: %v", err)
	}

	memo := types.Memo{
		From:    *from,
		To:      *to,
		Message: types.Buffer(msg),
		Nonce:   types.UInt64(nonce),
	}

	priv, err := types.NewPrivateKeyFromWif(wif)
	if err != nil {
		return "", fmt.Errorf("NewPrivateKeyFromWif: %v", err)
	}

	m := ""
	exception.ExceptionError{
		Try: func() {
			m, err = memo.Decrypt(priv)
			if err != nil {
				exception.Throw(fmt.Sprintf("decrypt error: %v", err))
			}
		},
		Catch: func(e exception.Exception) {
			err = fmt.Errorf("Unexpected %v", e)
		},
	}.Do()

	return m, err
}
