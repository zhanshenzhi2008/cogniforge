package skillimport

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// unzip 解压 ZIP 文件到目标目录
func unzip(srcPath, destDir string) error {
	reader, err := zip.OpenReader(srcPath)
	if err != nil {
		return fmt.Errorf("无法打开 ZIP 文件: %w", err)
	}
	defer reader.Close()

	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("无法创建解压目录: %w", err)
	}

	for _, file := range reader.File {
		// 安全检查：禁止路径穿越
		fpath := filepath.Clean(filepath.Join(destDir, file.Name))
		if !strings.HasPrefix(fpath, destDir+string(filepath.Separator)) {
			return fmt.Errorf("ZIP 包含非法路径: %s", file.Name)
		}

		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(fpath, 0755); err != nil {
				return err
			}
			continue
		}

		// 确保父目录存在
		if err := os.MkdirAll(filepath.Dir(fpath), 0755); err != nil {
			return err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			return fmt.Errorf("无法创建文件 %s: %w", fpath, err)
		}

		rc, err := file.Open()
		if err != nil {
			outFile.Close()
			return fmt.Errorf("无法读取 ZIP 条目 %s: %w", file.Name, err)
		}

		_, err = io.Copy(outFile, rc)
		rc.Close()
		outFile.Close()
		if err != nil {
			return fmt.Errorf("解压文件 %s 失败: %w", file.Name, err)
		}
	}
	return nil
}
