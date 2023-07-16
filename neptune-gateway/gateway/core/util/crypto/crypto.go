package crypto

import (
	"crypto/sha256"
	"fmt"
)

// GenerateSaltPassword 密码采用 SHA256 加盐加密
func GenerateSaltPassword(salt, password string) string {
	// 1. 采用 SHA256 加密
	firstHash := sha256.New()
	// 2. 哈希算法加密
	firstHash.Write([]byte(password))
	// 3. 计算哈希值
	hashCode := firstHash.Sum(nil)
	// 4. 加盐哈希 -> 不能直接采用 string() 转换字节数组
	secondHash := sha256.New()
	secondHash.Write([]byte(fmt.Sprintf("%x", hashCode) + salt))
	// 5. 计算哈希值
	return fmt.Sprintf("%x", secondHash.Sum(nil))
}
