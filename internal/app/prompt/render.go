package prompt

import (
	"fmt"
	"regexp"
	"strings"

	"cogniforge/internal/model"
)

var varPattern = regexp.MustCompile(`\{\{\s*([a-zA-Z_][\w]*)\s*\}\}`)

// ExtractVarNames 从正文提取 {{name}} 占位符（去重、保序）
func ExtractVarNames(content string) []string {
	seen := map[string]struct{}{}
	var out []string
	for _, m := range varPattern.FindAllStringSubmatch(content, -1) {
		if len(m) < 2 {
			continue
		}
		name := m[1]
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		out = append(out, name)
	}
	return out
}

// RenderContent 用 variables 替换 {{name}}；defs 提供 default / required 校验。
// missingRequired=true 时缺少必填变量则报错。
func RenderContent(content string, variables map[string]string, defs model.PromptVarDefs, missingRequired bool) (string, error) {
	defaults := map[string]string{}
	required := map[string]bool{}
	for _, d := range defs {
		name := strings.TrimSpace(d.Name)
		if name == "" {
			continue
		}
		defaults[name] = d.Default
		if d.Required {
			required[name] = true
		}
	}

	resolve := func(name string) (string, bool) {
		if variables != nil {
			if v, ok := variables[name]; ok {
				return v, true
			}
		}
		if v, ok := defaults[name]; ok && v != "" {
			return v, true
		}
		return "", false
	}

	if missingRequired {
		needed := ExtractVarNames(content)
		for _, name := range needed {
			if !required[name] {
				// 正文出现但未声明 required 的，不强制
				continue
			}
			if _, ok := resolve(name); !ok {
				return "", fmt.Errorf("缺少必填变量: %s", name)
			}
		}
		for name := range required {
			if _, ok := resolve(name); !ok {
				// 声明必填但正文未用到：仍要求传入，避免配置遗漏
				return "", fmt.Errorf("缺少必填变量: %s", name)
			}
		}
	}

	out := varPattern.ReplaceAllStringFunc(content, func(m string) string {
		sub := varPattern.FindStringSubmatch(m)
		if len(sub) < 2 {
			return m
		}
		name := sub[1]
		if v, ok := resolve(name); ok {
			return v
		}
		return m // 未提供则保留占位符，便于排查
	})
	return out, nil
}
