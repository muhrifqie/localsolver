package util

import (
	"crypto/rand"
	"encoding/binary"
	"math/big"
)

const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"

func RandStr(n int) string {
	b := make([]byte, n)
	for i := range b {
		idx, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		b[i] = charset[idx.Int64()]
	}
	return string(b)
}

func WidgetID() string {
	return "w_" + RandStr(12)
}

func AgentID() string {
	return "a_" + RandStr(12)
}

func SessionID() string {
	return RandStr(12)
}

func RandNonce() uint32 {
	var b [4]byte
	rand.Read(b[:])
	return binary.LittleEndian.Uint32(b[:])
}
