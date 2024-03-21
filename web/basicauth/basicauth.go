package basicauth

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"

	"golang.org/x/crypto/scrypt"

	"github.com/iotaledger/hive.go/ierrors"
	"github.com/iotaledger/hive.go/lo"
)

// SaltGenerator generates a crypto-secure random salt.
func SaltGenerator(length int) ([]byte, error) {
	salt := make([]byte, length)

	if _, err := rand.Read(salt); err != nil {
		return nil, err
	}

	return salt, nil
}

// DerivePasswordKey calculates the key based on password and salt.
func DerivePasswordKey(password []byte, salt []byte) ([]byte, error) {
	dk, err := scrypt.Key(password, salt, 1<<15, 8, 1, 32)
	if err != nil {
		return nil, err
	}

	return dk, err
}

// VerifyPassword verifies if the password is correct.
func VerifyPassword(password []byte, salt []byte, storedPasswordKey []byte) (bool, error) {
	dk, err := DerivePasswordKey(password, salt)
	if err != nil {
		return false, err
	}

	return bytes.Equal(dk, storedPasswordKey), nil
}

// BasicAuth is a basic authentication implementation for a single user.
//

type BasicAuth struct {
	username     string
	passwordHash []byte
	passwordSalt []byte
}

func NewBasicAuth(username string, passwordHashHex string, passwordSaltHex string) (*BasicAuth, error) {
	if len(username) == 0 {
		return nil, ierrors.New("username must not be empty")
	}

	passwordHash, err := decodePasswordHash(passwordHashHex)
	if err != nil {
		return nil, err
	}

	passwordSalt, err := decodePasswordSalt(passwordSaltHex)
	if err != nil {
		return nil, err
	}

	return &BasicAuth{
		username:     username,
		passwordHash: passwordHash,
		passwordSalt: passwordSalt,
	}, nil
}

func (b *BasicAuth) VerifyUsernameAndPasswordBytes(username string, passwordBytes []byte) bool {
	if username != b.username {
		return false
	}

	// error is ignored because it returns false in case it can't be derived
	return lo.Return1(VerifyPassword(passwordBytes, b.passwordSalt, b.passwordHash))
}

func (b *BasicAuth) VerifyUsernameAndPassword(username string, password string) bool {
	return b.VerifyUsernameAndPasswordBytes(username, []byte(password))
}

// BasicAuthManager is the same as BasicAuth but for multiple users.
//
//nolint:revive // better be explicit here
type BasicAuthManager struct {
	usersWithHashedPasswords map[string][]byte
	passwordSalt             []byte
}

func NewBasicAuthManager(usersWithPasswordsHex map[string]string, passwordSaltHex string) (*BasicAuthManager, error) {
	usersWithHashedPasswords := make(map[string][]byte, len(usersWithPasswordsHex))
	for username, passwordHashHex := range usersWithPasswordsHex {
		if len(username) == 0 {
			return nil, ierrors.New("username must not be empty")
		}

		password, err := decodePasswordHash(passwordHashHex)
		if err != nil {
			return nil, ierrors.Errorf("parsing password hash for user %s failed: %w", username, err)
		}
		usersWithHashedPasswords[username] = password
	}

	passwordSalt, err := decodePasswordSalt(passwordSaltHex)
	if err != nil {
		return nil, err
	}

	return &BasicAuthManager{
		usersWithHashedPasswords: usersWithHashedPasswords,
		passwordSalt:             passwordSalt,
	}, nil
}

func (b *BasicAuthManager) Exists(username string) bool {
	_, exists := b.usersWithHashedPasswords[username]
	return exists
}

func (b *BasicAuthManager) VerifyUsernameAndPasswordBytes(username string, passwordBytes []byte) bool {
	passwordHash, exists := b.usersWithHashedPasswords[username]
	if !exists {
		return false
	}

	// error is ignored because it returns false in case it can't be derived
	return lo.Return1(VerifyPassword(passwordBytes, b.passwordSalt, passwordHash))
}

func (b *BasicAuthManager) VerifyUsernameAndPassword(username string, password string) bool {
	return b.VerifyUsernameAndPasswordBytes(username, []byte(password))
}

func decodePasswordSalt(passwordSaltHex string) ([]byte, error) {
	if len(passwordSaltHex) != 64 {
		return nil, ierrors.New("password salt must be 64 (hex encoded) in length")
	}

	var err error
	passwordSalt, err := hex.DecodeString(passwordSaltHex)
	if err != nil {
		return nil, ierrors.Wrap(err, "password salt must be hex encoded")
	}

	return passwordSalt, nil
}

func decodePasswordHash(passwordHashHex string) ([]byte, error) {
	if len(passwordHashHex) != 64 {
		return nil, ierrors.New("password hash must be 64 (hex encoded scrypt hash) in length")
	}

	var err error
	passwordHash, err := hex.DecodeString(passwordHashHex)
	if err != nil {
		return nil, ierrors.Wrap(err, "password hash must be hex encoded")
	}

	return passwordHash, nil
}
