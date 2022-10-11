package util

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

const alphabet = "abcdefghijklmnopqrstuvwxyz"

func init() {
	rand.Seed(time.Now().UnixNano())
}

func RandomInt(min, max int64) int64 {
	return min + rand.Int63n(max-min+1)
}

func RandomString(n int) string {
	var sb strings.Builder
	k := len(alphabet)

	for i := 0; i < n; i++ {
		c := alphabet[rand.Intn(k)]
		sb.WriteByte(c)
	}
	return sb.String()
}

func RandomEmail() string {
	return fmt.Sprintf("%s@email.com", RandomString(6))
}

func RandomText(n int) string {
	var sb strings.Builder
	k := len(alphabet)
	const space = " "

	for i := 0; i < n; i++ {
		if i > 0 {
			sb.WriteByte(space[0])
		}
		ws := RandomInt(3, 12)
		for j := 0; j < int(ws); j++ {
			c := alphabet[rand.Intn(k)]
			sb.WriteByte(c)
		}
	}
	return sb.String()
}

const alphanumeric = "abcdefghijklmnopqrstuvwxyz1234567890"

func RandomNumString(n int) string {
	var sb strings.Builder
	k := len(alphanumeric)

	for i := 0; i < n; i++ {
		c := alphanumeric[rand.Intn(k)]
		sb.WriteByte(c)
	}
	return sb.String()
}

func RandomImageName() string {
	return fmt.Sprintf("img-%d-%s", time.Now().UnixNano()/1e6, RandomNumString(12))
}
