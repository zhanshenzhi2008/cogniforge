package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("🔍 GoLand 环境变量检查")
	fmt.Println("========================")

	keys := []string{
		"SMTP_HOST", "SMTP_PORT", "SMTP_USER", "SMTP_PASSWORD",
		"MAIL_FROM", "MAIL_PROVIDER",
		"PGSQL_HOST", "PGSQL_PORT", "REDIS_HOST",
	}

	allSet := true
	for _, k := range keys {
		v := os.Getenv(k)
		if v != "" {
			// 隐藏敏感信息
			if k == "SMTP_PASSWORD" || k == "PGSQL_PASSWORD" {
				if len(v) > 4 {
					fmt.Printf("✅ %s = %s***%s\n", k, v[:4], v[len(v)-4:])
				} else {
					fmt.Printf("✅ %s = [已设置]\n", k)
				}
			} else {
				fmt.Printf("✅ %s = %s\n", k, v)
			}
		} else {
			fmt.Printf("❌ %s = [未设置]\n", k)
			allSet = false
		}
	}

	fmt.Println("========================")
	if allSet {
		fmt.Println("✅ 所有配置已加载")
	} else {
		fmt.Println("❌ 有环境变量未设置！")
		fmt.Println("\n请在 GoLand Run Configuration 的 Environment 中添加：")
		fmt.Println("SMTP_HOST=smtp.qq.com")
		fmt.Println("SMTP_PORT=465")
		fmt.Println("SMTP_USER=zhanshenzhi2008@foxmail.com")
		fmt.Println("SMTP_PASSWORD=你的授权码")
		fmt.Println("MAIL_FROM=CogniForge <zhanshenzhi2008@foxmail.com>")
		fmt.Println("MAIL_PROVIDER=smtp")
	}
}
