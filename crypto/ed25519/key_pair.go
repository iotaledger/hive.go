package ed25519

type KeyPair struct {
	PrivateKey PrivateKey
	PublicKey  PublicKey
}

func GenerateKeyPair() (keyPair KeyPair) {
	public, private, err := GenerateKey()
	if err != nil {
		panic(err)
	}

	keyPair.PublicKey = public
	keyPair.PrivateKey = private

	return keyPair
}
