package keystore

type keyStorePassphrase struct {
	keysDirPath string
	scryptN     int
	scryptP     int
}
