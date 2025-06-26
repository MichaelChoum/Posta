/*
注意：
测试文件名必须以 _test.go 结尾
测试函数名必须以 Test 开头
测试文件可以没有 main 函数
运行测试命令：
go test -v ./pkg/encrypt: 指定测试目录，-v 表示详细输出
go test -v ./pkg/encrypt -run TestEncryptMobile： 运行指定的测试函数
go test：测试当前目录下所有的测试文件
*/

package encrypt

import (
	"testing"
)

func TestEncryptMobile(t *testing.T) {
	mobile := "13800138000"
	encryptedMobile, err := EncMobile(mobile)
	if err != nil {
		t.Fatal(err)
	}
	decryptedMobile, err := DecMobile(encryptedMobile)
	if err != nil {
		t.Fatal(err)
	}
	if mobile != decryptedMobile {
		t.Fatalf("expected %s, but got %s", mobile, decryptedMobile)
	}
}
