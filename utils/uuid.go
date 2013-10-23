/**
uuid.go : Ashish Banerjee, Sat, 12 May 12

Type 4 PRSG (Pseudo Random Sequence Geretaor) UUID

http://www.ietf.org/rfc/rfc4122.txt

*/
package utils

import (
	"encoding/hex"
	"crypto/rand"
)

func GenUUID() (string, error) {
	uuid := make([]byte, 16)
	n, err := rand.Read(uuid)
	if n != len(uuid) || err != nil {
		return "", err
	}
	// TODO: verify the two lines implement RFC 4122 correctly
	uuid[8] = 0x80 // variant bits see page 5
	uuid[4] = 0x40 // version 4 Pseudo Random, see page 7

	return hex.EncodeToString(uuid), nil
}