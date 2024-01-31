package modules

import (
	"crypto/sha256"
	"encoding/hex"
)

// Sha256Hash 加密转Hash
func Sha256Hash(input string) string {
	// SHA-256对象
	hasher := sha256.New()

	// 写入数据
	hasher.Write([]byte(input))

	// 计算哈希值
	hashInBytes := hasher.Sum(nil)

	// 转换十六进制字符串
	hashString := hex.EncodeToString(hashInBytes)

	return hashString
}
