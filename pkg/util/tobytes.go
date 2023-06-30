package util

import (
	"fmt"
)

func Stuff8Byte(b []byte) [8]byte {
	var res [8]byte
	if len(b) > 8 {
		b = b[0:9]
	}

	for index, _byte := range b {
		res[index] = _byte
	}
	return res
}

func Bytes2MacStr(b [8]byte) string {
	x := fmt.Sprintf("%x", b)
	x = x[:len(x)-4]
	res := ""
	for i := 0; i <= len(x)-2; i += 2 {
		res += x[i : i+2]
		if i+2 != len(x) {
			res += ":"
		}
	}
	return res
}
