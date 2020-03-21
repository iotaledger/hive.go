package ed25519

type KeyPair struct {
	PrivateKey PrivateKey
	PublicKey  PublicKey
}

func GenerateKeyPair() (keyPair KeyPair) {
	if public, private, err := GenerateKey(); err != nil {
		panic(err)
	} else {
		keyPair.PublicKey = public
		keyPair.PrivateKey = private

		return
	}
}
