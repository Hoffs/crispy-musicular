package rand

import (
	"crypto/rand"
	"math/big"
)

var charset = []byte("0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz")
var charsetLen = big.NewInt(int64(len(charset)))

func String(n uint8) (string, error) {
	d := make([]byte, n)
	for id := range d {
		r, err := rand.Int(rand.Reader, charsetLen)
		if err != nil {
			return "", err
		}
		d[id] = charset[r.Int64()]
	}
	rand.Int(rand.Reader, charsetLen)

	return string(d), nil
}
