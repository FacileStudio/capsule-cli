package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/pbkdf2"
)

func ToBase64URL(data []byte) string {
	return base64.RawURLEncoding.EncodeToString(data)
}

func FromBase64URL(s string) ([]byte, error) {
	return base64.RawURLEncoding.DecodeString(s)
}

func GenerateKey() ([]byte, error) {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return nil, fmt.Errorf("generating key: %w", err)
	}
	return key, nil
}

func Encrypt(plaintext []byte, key []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("creating cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("creating GCM: %w", err)
	}

	iv := make([]byte, 12)
	if _, err := rand.Read(iv); err != nil {
		return "", fmt.Errorf("generating IV: %w", err)
	}

	ciphertext := gcm.Seal(nil, iv, plaintext, nil)
	combined := append(iv, ciphertext...)
	return base64.StdEncoding.EncodeToString(combined), nil
}

func Decrypt(data string, key []byte) ([]byte, error) {
	raw, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, fmt.Errorf("decoding base64: %w", err)
	}

	if len(raw) < 12 {
		return nil, fmt.Errorf("ciphertext too short")
	}

	iv := raw[:12]
	ciphertext := raw[12:]

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("creating cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("creating GCM: %w", err)
	}

	plaintext, err := gcm.Open(nil, iv, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("decrypting: %w", err)
	}

	return plaintext, nil
}

func WrapKey(contentKey []byte, password string) (string, error) {
	salt := make([]byte, 32)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("generating salt: %w", err)
	}

	wrappingKey := pbkdf2.Key([]byte(password), salt, 600000, 32, sha256.New)

	iv := make([]byte, 12)
	if _, err := rand.Read(iv); err != nil {
		return "", fmt.Errorf("generating IV: %w", err)
	}

	block, err := aes.NewCipher(wrappingKey)
	if err != nil {
		return "", fmt.Errorf("creating cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("creating GCM: %w", err)
	}

	encryptedKey := gcm.Seal(nil, iv, contentKey, nil)

	fragment := ToBase64URL(encryptedKey) + "." + ToBase64URL(salt) + "." + ToBase64URL(iv)
	return fragment, nil
}

func UnwrapKey(fragment string, password string) ([]byte, error) {
	parts := strings.Split(fragment, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid password-protected fragment format")
	}

	encryptedKey, err := FromBase64URL(parts[0])
	if err != nil {
		return nil, fmt.Errorf("decoding encrypted key: %w", err)
	}

	salt, err := FromBase64URL(parts[1])
	if err != nil {
		return nil, fmt.Errorf("decoding salt: %w", err)
	}

	iv, err := FromBase64URL(parts[2])
	if err != nil {
		return nil, fmt.Errorf("decoding IV: %w", err)
	}

	wrappingKey := pbkdf2.Key([]byte(password), salt, 600000, 32, sha256.New)

	block, err := aes.NewCipher(wrappingKey)
	if err != nil {
		return nil, fmt.Errorf("creating cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("creating GCM: %w", err)
	}

	contentKey, err := gcm.Open(nil, iv, encryptedKey, nil)
	if err != nil {
		return nil, fmt.Errorf("unwrapping key (wrong password?): %w", err)
	}

	return contentKey, nil
}

func IsPasswordProtected(fragment string) bool {
	return strings.Contains(fragment, ".")
}
