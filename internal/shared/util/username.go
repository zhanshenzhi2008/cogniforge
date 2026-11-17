package util

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

// 用户名：2–20 字；字母/数字/下划线/连字符/中文；禁止空格及其它空白。
var usernamePattern = regexp.MustCompile(`^[\p{L}\p{N}_-]{2,20}$`)

// NormalizeUsername 去掉首尾空白。
func NormalizeUsername(s string) string {
	return strings.TrimSpace(s)
}

// ValidateUsername 校验登录用户名格式（调用前建议先 NormalizeUsername）。
func ValidateUsername(name string) error {
	if name == "" {
		return fmt.Errorf("请输入用户名")
	}
	for _, r := range name {
		if unicode.IsSpace(r) {
			return fmt.Errorf("用户名不能包含空格")
		}
	}
	n := utf8.RuneCountInString(name)
	if n < 2 || n > 20 {
		return fmt.Errorf("用户名长度为 2–20 个字符")
	}
	if !usernamePattern.MatchString(name) {
		return fmt.Errorf("用户名只能包含字母、数字、下划线、连字符或中文")
	}
	return nil
}
