package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateUsername(t *testing.T) {
	assert.NoError(t, ValidateUsername("admin"))
	assert.NoError(t, ValidateUsername("张三"))
	assert.NoError(t, ValidateUsername("user_01"))
	assert.NoError(t, ValidateUsername("a-b"))

	assert.EqualError(t, ValidateUsername(""), "请输入用户名")
	assert.EqualError(t, ValidateUsername("a"), "用户名长度为 2–20 个字符")
	assert.EqualError(t, ValidateUsername("Test User"), "用户名不能包含空格")
	assert.EqualError(t, ValidateUsername("ab cd"), "用户名不能包含空格")
	assert.EqualError(t, ValidateUsername("bad!"), "用户名只能包含字母、数字、下划线、连字符或中文")
}
