package encoding

import (
	"fmt"

	"github.com/denkhaus/bitshares/config"
	"github.com/denkhaus/bitshares/types"
)

//Decrypt calculates a shared secret by the receivers private key
//and the senders public key, then decrypts the given memo message.
func Decrypt(msg, pub, wif string) (string, error) {
	var buf types.Buffer

	config.SetCurrent(config.ChainIDBTS)
	from, _ := types.NewPublicKeyFromString(pub)
	buf.FromString(msg)

	memo := types.Memo{
		From:    *from,
		Message: buf,
	}

	priv, err := types.NewPrivateKeyFromWif(wif)
	if err != nil {
		return "", fmt.Errorf("NewPrivateKeyFromWif", err)
	}

	m, err := memo.Decrypt(priv)
	if err != nil {
		return "", fmt.Errorf("Decrypt", err)
	}

	return m, nil
}
