package nodes

import (
	"fmt"
	"strings"

	"cogniforge/internal/engine/core"
)

// resolveTemplate 替换 {{key}}：优先 Variables，其次 Input 同名键。
func resolveTemplate(template string, ctx *core.ExecutionContext) string {
	if template == "" || ctx == nil {
		return template
	}
	out := template
	// 简单扫描：对 Variables 与 Input 的每个 string/可格式化值做替换
	replaceMap := map[string]string{}
	for k, v := range ctx.Input {
		replaceMap[k] = fmt.Sprint(v)
	}
	for k, v := range ctx.Variables {
		replaceMap[k] = fmt.Sprint(v)
	}
	for k, v := range replaceMap {
		out = strings.ReplaceAll(out, "{{"+k+"}}", v)
	}
	return out
}
