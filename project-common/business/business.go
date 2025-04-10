package business

import (
	"fmt"
	"strconv"
)

func StringToInt32(s string) (int32, error) {
	// 使用 strconv.Atoi 将字符串转换为 int
	i, err := strconv.Atoi(s)
	if err != nil {
		return 0, err
	}

	// 检查转换后的值是否在 int32 的范围内
	if i < -2147483648 || i > 2147483647 {
		return 0, fmt.Errorf("value out of range for int32")
	}

	return int32(i), nil
}
