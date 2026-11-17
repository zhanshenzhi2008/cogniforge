package skillimport

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/viper"

	"cogniforge/internal/model"
)

// =============================================================================
// 安全配置
// =============================================================================

// URL 安全配置（从 config.yaml 读取）
type URLSecurityConfig struct {
	AllowedSchemes    []string // 允许的协议，默认 ["https"]
	BlockedHosts      []string // 拉黑的域名/IP
	AllowedHosts      []string // 信任域名白名单（可放行 HTTP）
	MaxImportSizeMB   int      // 最大导入大小 MB，默认 10
	MaxReferenceCount int      // 最大引用数，默认 50
}

func DefaultSecurityConfig() URLSecurityConfig {
	return URLSecurityConfig{
		AllowedSchemes:    []string{"https"},
		BlockedHosts:      []string{"localhost", "127.0.0.1", "0.0.0.0"},
		AllowedHosts:      []string{},
		MaxImportSizeMB:   10,
		MaxReferenceCount: 50,
	}
}

// =============================================================================
// 校验结果
// =============================================================================

// ValidationResult 校验结果
type ValidationResult struct {
	Valid   bool
	Errors  []string
	Warnings []string
}

func (v *ValidationResult) AddError(format string, args ...interface{}) {
	v.Valid = false
	v.Errors = append(v.Errors, fmt.Sprintf(format, args...))
}

func (v *ValidationResult) AddWarning(format string, args ...interface{}) {
	v.Warnings = append(v.Warnings, fmt.Sprintf(format, args...))
}

// =============================================================================
// URL 安全校验器
// =============================================================================

// URLValidator URL 安全校验器
type URLValidator struct {
	cfg URLSecurityConfig
}

func NewURLValidator(cfg URLSecurityConfig) *URLValidator {
	if len(cfg.AllowedSchemes) == 0 {
		cfg.AllowedSchemes = []string{"https"}
	}
	return &URLValidator{cfg: cfg}
}

// ValidateURL 校验单个 URL 是否安全
func (v *URLValidator) ValidateURL(rawURL string) error {
	if rawURL == "" {
		return nil
	}

	u, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("URL 格式错误: %s", rawURL)
	}

	// 1. 协议校验
	scheme := strings.ToLower(u.Scheme)
	allowed := false
	for _, s := range v.cfg.AllowedSchemes {
		if scheme == s {
			allowed = true
			break
		}
	}
	// 信任域名可放行 HTTP
	if !allowed && scheme == "http" {
		host := strings.ToLower(v.trimPort(u.Host))
		for _, h := range v.cfg.AllowedHosts {
			if host == strings.ToLower(h) {
				allowed = true
				break
			}
		}
	}
	if !allowed {
		return fmt.Errorf("不允许的协议: %s（仅支持 %v）", scheme, v.cfg.AllowedSchemes)
	}

	// 2. 危险协议
	lower := strings.ToLower(rawURL)
	if strings.HasPrefix(lower, "javascript:") ||
		strings.HasPrefix(lower, "data:") ||
		strings.HasPrefix(lower, "file:") ||
		strings.HasPrefix(lower, "vbscript:") {
		return fmt.Errorf("禁止的危险协议: %s", scheme)
	}

	// 3. 主机名校验
	host := v.trimPort(u.Host)
	if host == "" {
		return fmt.Errorf("URL 缺少主机名")
	}

	// 内网 IP 检测
	ip := net.ParseIP(host)
	if ip != nil {
		if v.isPrivateOrReservedIP(ip) {
			return fmt.Errorf("禁止访问内网 IP: %s", host)
		}
		return nil
	}

	// 主机名黑名单
	lowerHost := strings.ToLower(host)
	for _, blocked := range v.cfg.BlockedHosts {
		if lowerHost == strings.ToLower(blocked) {
			return fmt.Errorf("禁止访问: %s", host)
		}
	}

	// 解析域名并检查所有 IP
	ips, err := net.LookupIP(host)
	if err == nil {
		for _, ip := range ips {
			if v.isPrivateOrReservedIP(ip) {
				return fmt.Errorf("域名 %s 解析到内网 IP %s，已被阻断", host, ip.String())
			}
		}
	}

	return nil
}

// isPrivateOrReservedIP 判断是否为内网/保留 IP
// 参考 RFC 1918 / RFC 4193 / RFC 3927
func (v *URLValidator) isPrivateOrReservedIP(ip net.IP) bool {
	// 127.0.0.0/8  loopback
	// 10.0.0.0/8   私有
	// 172.16.0.0/12 私有
	// 192.168.0.0/16 私有
	// 169.254.0.0/16 link-local
	// 0.0.0.0/8
	private := []string{
		"127.", "10.", "172.16.", "172.17.", "172.18.", "172.19.",
		"172.20.", "172.21.", "172.22.", "172.23.", "172.24.", "172.25.",
		"172.26.", "172.27.", "172.28.", "172.29.", "172.30.", "172.31.",
		"192.168.", "169.254.", "0.",
	}
	ipStr := ip.String()
	for _, p := range private {
		if strings.HasPrefix(ipStr, p) {
			return true
		}
	}
	return false
}

func (v *URLValidator) trimPort(host string) string {
	if h, _, err := net.SplitHostPort(host); err == nil {
		return h
	}
	return host
}

// =============================================================================
// SKILL.md 格式解析器
// =============================================================================

// ClaudeSkillFrontmatter Claude SKILL.md YAML frontmatter
type ClaudeSkillFrontmatter struct {
	Name                  string   `yaml:"name"`
	Description           string   `yaml:"description"`
	WhenToUse             string   `yaml:"when_to_use"`
	AllowedTools          []string `yaml:"allowed-tools"`
	Arguments             string   `yaml:"arguments"`
	ArgumentsList         []string `yaml:"arguments_list"`
	Context               string   `yaml:"context"` // "fork"
	Model                 string   `yaml:"model"`
	Effort                string   `yaml:"effort"`
	Version               string   `yaml:"version"`
	UserInvocable         bool     `yaml:"user-invocable"`
	DisableModelInvocation bool    `yaml:"disable-model-invocation"`
	Paths                 string   `yaml:"paths"`
}

// ParsedSkill 解析后的 Skill
type ParsedSkill struct {
	Frontmatter ClaudeSkillFrontmatter
	Instructions string // markdown body
	Files       map[string][]byte // 附加文件: path -> content
}

// SkillParser SKILL.md 解析器
type SkillParser struct {
	urlValidator *URLValidator
}

func NewSkillParser(cfg URLSecurityConfig) *SkillParser {
	return &SkillParser{
		urlValidator: NewURLValidator(cfg),
	}
}

// ParseFromMarkdown 从 SKILL.md 文本解析
func (p *SkillParser) ParseFromMarkdown(content string) (*ParsedSkill, *ValidationResult) {
	result := &ValidationResult{Valid: true}
	ps := &ParsedSkill{
		Files: make(map[string][]byte),
	}

	// 拆 frontmatter 和 body
	frontmatter, body, found := strings.Cut(strings.TrimLeft(content, "\r\n"), "---")
	if !found || frontmatter == "" {
		result.AddError("缺少 YAML frontmatter（--- 分隔符）")
		return nil, result
	}
	frontmatter = strings.Trim(frontmatter, "\r\n")
	body = strings.Trim(strings.TrimPrefix(body, "---"), "\r\n")

	// 解析 YAML
	if err := yamlUnmarshal(frontmatter, &ps.Frontmatter); err != nil {
		result.AddError("YAML 解析失败: %v", err)
		return nil, result
	}

	// 必填字段
	if ps.Frontmatter.Name == "" {
		result.AddError("frontmatter 缺少 name 字段")
	}
	if ps.Frontmatter.Description == "" {
		result.AddError("frontmatter 缺少 description 字段")
	}

	ps.Instructions = body

	return ps, result
}

// ParseZipFromReader 从 ZIP 内容解析，返回多个 Skill
func (p *SkillParser) ParseZipFromReader(r io.Reader, size int64) ([]*ParsedSkill, *ValidationResult) {
	result := &ValidationResult{Valid: true}

	// 大小校验（前置，避免写大文件）
	maxSize := int64(p.urlValidator.cfg.MaxImportSizeMB) * 1024 * 1024
	if size > maxSize {
		result.AddError("导入文件超过大小限制（%d MB）", p.urlValidator.cfg.MaxImportSizeMB)
		return nil, result
	}

	// 写到临时目录
	tmpDir, err := os.MkdirTemp("", "skill-import-*")
	if err != nil {
		result.AddError("无法创建临时目录: %v", err)
		return nil, result
	}
	defer os.RemoveAll(tmpDir)

	zipPath := filepath.Join(tmpDir, "import.zip")
	f, err := os.Create(zipPath)
	if err != nil {
		result.AddError("无法创建临时文件: %v", err)
		return nil, result
	}
	if _, err := io.Copy(f, r); err != nil {
		f.Close()
		result.AddError("写入临时文件失败: %v", err)
		return nil, result
	}
	f.Close()

	extractDir := filepath.Join(tmpDir, "extracted")
	if err := unzip(zipPath, extractDir); err != nil {
		result.AddError("ZIP 解压失败: %v", err)
		return nil, result
	}

	// 查找所有 SKILL.md（根目录单个 或 子目录多个）
	var skillFiles []string
	knownDirs := map[string]bool{"references": true, "scripts": true, "assets": true, "templates": true}

	_ = filepath.Walk(extractDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		base := filepath.Base(path)
		if base == "SKILL.md" {
			skillFiles = append(skillFiles, path)
		}
		return nil
	})

	if len(skillFiles) == 0 {
		result.AddError("压缩包中未找到任何 SKILL.md 文件")
		return nil, result
	}

	var skills []*ParsedSkill
	for _, smPath := range skillFiles {
		sm, err := os.ReadFile(smPath)
		if err != nil {
			result.AddWarning("无法读取 %s: %v，跳过", smPath, err)
			continue
		}

		ps, parseResult := p.ParseFromMarkdown(string(sm))
		if !parseResult.Valid {
			result.AddWarning("解析 %s 失败: %v，跳过", smPath, parseResult.Errors)
			continue
		}

		// 收集该 SKILL 所在目录的附加文件
		skillDir := filepath.Dir(smPath)
		_ = filepath.Walk(skillDir, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() || path == smPath {
				return nil
			}
			rel, _ := filepath.Rel(skillDir, path)
			parts := strings.SplitN(rel, string(filepath.Separator), 2)
			if len(parts) >= 1 && knownDirs[parts[0]] {
				data, _ := os.ReadFile(path)
				if data != nil {
					ps.Files[rel] = data
				}
			}
			return nil
		})

		skills = append(skills, ps)
	}

	if len(skills) == 0 {
		result.AddError("没有找到可导入的 SKILL（所有文件解析均失败）")
		return nil, result
	}

	return skills, result
}

// ValidateParsedSkill 校验解析后的 Skill 内容
func (p *SkillParser) ValidateParsedSkill(ps *ParsedSkill, userID string) *ValidationResult {
	result := &ValidationResult{Valid: true}

	// 字段长度校验
	if len(ps.Frontmatter.Name) > 255 {
		result.AddError("name 字段超过 255 字符")
	}
	if len(ps.Frontmatter.Description) > 1536 {
		result.AddWarning("description 超过 1536 字符，可能被截断")
	}
	if len(ps.Instructions) > 50000 {
		result.AddError("instructions 超过 50000 字符")
	}

	// context=fork 警告（沙箱隔离未实现）
	if ps.Frontmatter.Context == "fork" {
		result.AddWarning("context: fork 需要子代理支持，当前未完整实现")
	}

	return result
}

// ToSkillModel 转换为 model.Skill
func (p *SkillParser) ToSkillModel(ps *ParsedSkill, userID string) *model.Skill {
	fm := ps.Frontmatter
	id := fmt.Sprintf("skill_%s", hashString(fm.Name))

	// 将 allowed-tools 转为 constraints
	var constraints model.JSONBArray
	for _, tool := range fm.AllowedTools {
		constraints = append(constraints, "允许使用: "+tool)
	}

	skill := &model.Skill{
		ID:           id,
		UserID:       userID,
		Name:         fm.Name,
		Description:  fm.Description,
		Instructions: ps.Instructions,
		Icon:         "🛠️",
		Examples:     model.JSONBArray{},
		References:   model.JSONBArray{},
		Constraints:  constraints,
		Model:        fm.Model,
		McpServers:   model.JSONBArray{},
		MemoryType:   "short_term",
		MemoryTurns:  10,
		Category:     "imported",
		Tags:         model.JSONBArray{},
		Version:      fm.Version,
		Author:       "imported",
		IsBuiltIn:    false,
		SortOrder:    0,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// when_to_use 追加到 description
	if fm.WhenToUse != "" {
		skill.Description += "\n\n适用场景：" + fm.WhenToUse
	}

	return skill
}

// =============================================================================
// 工具函数
// =============================================================================

func hashString(s string) string {
	h := sha256.Sum256([]byte(s + time.Now().Format("20060102")))
	return hex.EncodeToString(h[:8])
}

// yamlUnmarshal 简单 YAML 解析（不依赖第三方库）
func yamlUnmarshal(content string, out interface{}) error {
	lines := strings.Split(content, "\n")
	data := make(map[string]interface{})

	for len(lines) > 0 && strings.TrimSpace(lines[0]) == "" {
		lines = lines[1:]
	}
	for len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "" {
		lines = lines[:len(lines)-1]
	}

	var currentKey string
	var listItems []string
	inList := false

	for _, line := range lines {
		trimmed := strings.TrimLeft(line, " \t")
		if trimmed == "" {
			continue
		}

		// 列表项
		if strings.HasPrefix(trimmed, "- ") {
			val := strings.TrimSpace(strings.TrimPrefix(trimmed, "- "))
			if currentKey != "" && !inList {
				data[currentKey] = strings.Join(listItems, ",")
				listItems = nil
			}
			listItems = append(listItems, val)
			inList = true
			continue
		}

		// 保存前一个列表
		if currentKey != "" && inList {
			data[currentKey] = listItems
			listItems = nil
			inList = false
		}

		// 键值对
		parts := strings.SplitN(trimmed, ":", 2)
		if len(parts) == 2 {
			currentKey = strings.TrimSpace(parts[0])
			val := strings.TrimSpace(parts[1])
			if val == "" {
				inList = false
				continue
			}
			// 去掉引号
			val = strings.Trim(val, "\"")
			val = strings.Trim(val, "'")
			data[currentKey] = val
			currentKey = ""
		}
	}
	if currentKey != "" && inList {
		data[currentKey] = listItems
	}

	// 映射到结构体
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return json.Unmarshal(jsonData, out)
}

// ValidateMultipartFile 校验上传文件
func ValidateMultipartFile(fh *multipart.FileHeader, cfg URLSecurityConfig) *ValidationResult {
	result := &ValidationResult{Valid: true}
	maxSize := int64(cfg.MaxImportSizeMB) * 1024 * 1024

	if fh.Size > maxSize {
		result.AddError("文件大小 %d MB 超过限制 %d MB", fh.Size/1024/1024, cfg.MaxImportSizeMB)
	}

	ext := strings.ToLower(filepath.Ext(fh.Filename))
	allowedExts := map[string]bool{".md": true, ".zip": true}
	if !allowedExts[ext] {
		result.AddError("不支持的文件类型: %s（仅支持 .md, .zip）", ext)
	}

	return result
}

// LoadSecurityConfig 尝试从 viper 加载配置
func LoadSecurityConfig() URLSecurityConfig {
	cfg := DefaultSecurityConfig()

	if runtime.GOMAXPROCS(0) > 0 && viper.GetViper() != nil {
		v := viper.GetViper()
		if v.GetInt("import.max_size_mb") > 0 {
			cfg.MaxImportSizeMB = v.GetInt("import.max_size_mb")
		}
		if v.GetStringSlice("import.allowed_hosts") != nil {
			cfg.AllowedHosts = v.GetStringSlice("import.allowed_hosts")
		}
	}
	return cfg
}
