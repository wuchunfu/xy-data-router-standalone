package internal

import (
	"strconv"

	"github.com/fufuok/utils"
)

// HashString 合并一串文本, 得到字符串哈希
func HashString(s ...string) string {
	return strconv.FormatUint(HashStringUint64(s...), 10)
}

func HashStringUint64(s ...string) uint64 {
	return utils.Sum64(utils.AddString(s...))
}
