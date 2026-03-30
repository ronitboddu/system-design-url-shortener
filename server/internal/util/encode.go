package util

import (
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"io"
	"strings"
)

type Encode interface {
	toBase62(n uint64)
}

const base62Chars = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func toBase62(n uint64) string {
	if n == 0 {
		return "0"
	}

	var result []byte
	for n > 0 {
		result = append([]byte{base62Chars[n%62]}, result...)
		n /= 62
	}
	return string(result)
}

func GetCode(ip_url string) string {
	h := md5.New()
	io.WriteString(h, ip_url)
	hash_value := fmt.Sprintf("%x", h.Sum(nil))
	num := binary.BigEndian.Uint64([]byte(hash_value[:8]))

	// base 62 encode
	code := toBase62(num)
	if len(code) < 7 {
		code = strings.Repeat("0", 7-len(code)) + code
	}

	return code[:7]
}
