package notify

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hkdf"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
)

// ErrSecretUnreadable means a stored config could not be decrypted (wrong key,
// tampered blob, or a row swapped to another channel id).
var ErrSecretUnreadable = errors.New("notify: channel secret unreadable")

const (
	keyIDExplicit = 1 // NOTIFY_SECRET_KEY
	keyIDDerived  = 2 // HKDF(SESSION_SECRET)
	hkdfInfo      = "kith/notify/v1"
)

// SecretBox encrypts channel configs with AES-256-GCM.
// Blob layout: keyID(1) || nonce(12) || ciphertext. AAD binds the blob to its row.
type SecretBox struct {
	keys    map[byte][]byte
	sealing byte
}

// NewSecretBox takes the optional base64 NOTIFY_SECRET_KEY (32 bytes) and the
// session secret. Both keys stay available for decryption, so setting an
// explicit key later does not orphan rows sealed with the derived one.
func NewSecretBox(notifyKey, sessionSecret string) (*SecretBox, error) {
	b := &SecretBox{keys: map[byte][]byte{}}

	if sessionSecret != "" {
		k, err := hkdf.Key(sha256.New, []byte(sessionSecret), nil, hkdfInfo, 32)
		if err != nil {
			return nil, fmt.Errorf("notify: derive key: %w", err)
		}

		b.keys[keyIDDerived] = k
		b.sealing = keyIDDerived
	}

	if notifyKey != "" {
		k, err := base64.StdEncoding.DecodeString(notifyKey)
		if err != nil || len(k) != 32 {
			return nil, errors.New("notify: NOTIFY_SECRET_KEY must be 32 bytes, base64 encoded")
		}

		b.keys[keyIDExplicit] = k
		b.sealing = keyIDExplicit
	}

	if b.sealing == 0 {
		return nil, errors.New("notify: no secret key available (set NOTIFY_SECRET_KEY or SESSION_SECRET)")
	}

	return b, nil
}

func newGCM(key []byte) (cipher.AEAD, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	return cipher.NewGCM(block)
}

func (b *SecretBox) Seal(aad string, plaintext []byte) ([]byte, error) {
	gcm, err := newGCM(b.keys[b.sealing])
	if err != nil {
		return nil, fmt.Errorf("notify: seal: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, fmt.Errorf("notify: seal nonce: %w", err)
	}

	out := append([]byte{b.sealing}, nonce...)

	return gcm.Seal(out, nonce, plaintext, []byte(aad)), nil
}

func (b *SecretBox) Open(aad string, blob []byte) ([]byte, error) {
	if len(blob) < 1 {
		return nil, ErrSecretUnreadable
	}

	key, ok := b.keys[blob[0]]
	if !ok {
		return nil, ErrSecretUnreadable
	}

	gcm, err := newGCM(key)
	if err != nil || len(blob) < 1+gcm.NonceSize() {
		return nil, ErrSecretUnreadable
	}

	nonce, ct := blob[1:1+gcm.NonceSize()], blob[1+gcm.NonceSize():]

	pt, err := gcm.Open(nil, nonce, ct, []byte(aad))
	if err != nil {
		return nil, ErrSecretUnreadable
	}

	return pt, nil
}
