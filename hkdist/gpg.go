package main

import (
	"code.google.com/p/go.crypto/openpgp"
	"os"
)

var pubKeyRing openpgp.EntityList

func PubKeyRing() (openpgp.EntityList, error) {
	if pubKeyRing == nil {
		pubringFile, err := os.Open("pubring.gpg")
		if err != nil {
			return nil, err
		}
		defer pubringFile.Close()

		pubKeyRing, err = openpgp.ReadKeyRing(pubringFile)
		if err != nil {
			return nil, err
		}
	}
	return pubKeyRing, nil
}
