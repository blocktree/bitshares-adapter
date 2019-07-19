package txsigner

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"log"

	"github.com/blocktree/bitshares-adapter/encoding"
	"github.com/blocktree/bitshares-adapter/types"

	"github.com/pkg/errors"
)

type SignedTransaction struct {
	*types.Transaction
}

func NewSignedTransaction(tx *types.Transaction) *SignedTransaction {
	return &SignedTransaction{tx}
}

func (tx *SignedTransaction) Serialize() ([]byte, error) {
	var b bytes.Buffer
	encoder := encoding.NewEncoder(&b)

	if err := encoder.Encode(tx.Transaction); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func (tx *SignedTransaction) ID() (string, error) {
	var msgBuffer bytes.Buffer

	// Write the serialized transaction.
	rawTx, err := tx.Serialize()
	if err != nil {
		return "", err
	}

	if _, err := msgBuffer.Write(rawTx); err != nil {
		return "", errors.Wrap(err, "failed to write serialized transaction")
	}

	msgBytes := msgBuffer.Bytes()

	// Compute the digest.
	digest := sha256.Sum256(msgBytes)

	id := hex.EncodeToString(digest[:])
	length := 40
	if len(id) < 40 {
		length = len(id)
	}
	return id[:length], nil
}

func (tx *SignedTransaction) Digest(chain string) ([]byte, error) {
	var msgBuffer bytes.Buffer

	// Write the chain ID.
	rawChainID, err := hex.DecodeString(chain)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to decode chain ID: %v", chain)
	}

	if _, err := msgBuffer.Write(rawChainID); err != nil {
		return nil, errors.Wrap(err, "failed to write chain ID")
	}

	// Write the serialized transaction.
	rawTx, err := tx.Serialize()
	if err != nil {
		return nil, err
	}

	if _, err := msgBuffer.Write(rawTx); err != nil {
		return nil, errors.Wrap(err, "failed to write serialized transaction")
	}

	msgBytes := msgBuffer.Bytes()
	message := hex.EncodeToString(msgBytes)
	log.Printf("[DEBUG] Digest final message:%s\n", message)

	// Compute the digest.
	digest := sha256.Sum256(msgBytes)
	return digest[:], nil
}
