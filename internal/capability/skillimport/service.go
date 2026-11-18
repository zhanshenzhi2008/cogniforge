package skillimport

import (
	"io"
	"regexp"
	"strings"

	"cogniforge/internal/model"
	"gorm.io/gorm"
)

// urlRe 预编译 URL 正则
var urlRe = regexp.MustCompile(`https?://[^\s<>"{}|\\^` + "`" + `\[\]]+`)

// Service 导入服务
type Service struct {
	db           *gorm.DB
	parser       *SkillParser
	urlValidator *URLValidator
}

// NewService 创建 Import Service
func NewService(db *gorm.DB) *Service {
	cfg := LoadSecurityConfig()
	parser := NewSkillParser(cfg)
	urlValidator := NewURLValidator(cfg)

	return &Service{
		db:           db,
		parser:       parser,
		urlValidator: urlValidator,
	}
}

// =============================================================================
// DTO
// =============================================================================

// ImportResult 导入结果（单个）
type ImportResult struct {
	Skill    *model.Skill `json:"skill,omitempty"`
	Errors   []string     `json:"errors,omitempty"`
	Warnings []string     `json:"warnings,omitempty"`
}

// BatchImportResult 批量导入结果
type BatchImportResult struct {
	Total   int             `json:"total"`    // 尝试导入数量
	Success int             `json:"success"`   // 成功数量
	Failed  int             `json:"failed"`    // 失败数量
	Results []ImportResult  `json:"results"`   // 每项结果
}

// =============================================================================
// Skill 导入
// =============================================================================

// ImportSkillFromMarkdown 从 Markdown 文本导入
func (s *Service) ImportSkillFromMarkdown(userID string, content []byte) *ImportResult {
	result := &ImportResult{}

	// 1. 解析 SKILL.md
	ps, parseResult := s.parser.ParseFromMarkdown(string(content))
	if !parseResult.Valid {
		result.Errors = parseResult.Errors
		return result
	}
	result.Warnings = append(result.Warnings, parseResult.Warnings...)

	// 2. 校验内容
	validResult := s.parser.ValidateParsedSkill(ps, userID)
	if !validResult.Valid {
		result.Errors = validResult.Errors
		return result
	}
	result.Warnings = append(result.Warnings, validResult.Warnings...)

	// 3. 校验引用中的 URL
	refResult := s.validateReferences(ps.Instructions)
	if len(refResult.Errors) > 0 {
		result.Errors = refResult.Errors
		return result
	}
	result.Warnings = append(result.Warnings, refResult.Warnings...)

	// 4. 转换并保存
	skill := s.parser.ToSkillModel(ps, userID)
	if err := s.db.Create(skill).Error; err != nil {
		result.Errors = []string{"保存失败: " + err.Error()}
		return result
	}

	result.Skill = skill
	return result
}

// ImportSkillFromZIP 从 ZIP 文件导入（支持批量）
func (s *Service) ImportSkillFromZIP(userID string, r io.Reader, size int64) *BatchImportResult {
	result := &BatchImportResult{}

	// 1. 解析 ZIP（返回多个 Skill）
	skills, parseResult := s.parser.ParseZipFromReader(r, size)
	if skills == nil {
		result.Results = append(result.Results, ImportResult{Errors: parseResult.Errors})
		result.Failed = 1
		return result
	}

	result.Total = len(skills)
	for _, ps := range skills {
		item := ImportResult{}

		// 2. 校验内容
		validResult := s.parser.ValidateParsedSkill(ps, userID)
		if !validResult.Valid {
			item.Errors = validResult.Errors
			result.Results = append(result.Results, item)
			result.Failed++
			continue
		}
		item.Warnings = append(item.Warnings, validResult.Warnings...)

		// 3. 校验正文中所有 URL
		refResult := s.validateReferences(ps.Instructions)
		for _, data := range ps.Files {
			refResult2 := s.validateReferences(string(data))
			refResult.Warnings = append(refResult.Warnings, refResult2.Warnings...)
		}
		if len(refResult.Errors) > 0 {
			item.Errors = refResult.Errors
			result.Results = append(result.Results, item)
			result.Failed++
			continue
		}
		item.Warnings = append(item.Warnings, refResult.Warnings...)

		// 4. 转换并保存
		skill := s.parser.ToSkillModel(ps, userID)
		if err := s.db.Create(skill).Error; err != nil {
			item.Errors = []string{"保存失败: " + err.Error()}
			result.Results = append(result.Results, item)
			result.Failed++
			continue
		}

		item.Skill = skill
		result.Results = append(result.Results, item)
		result.Success++
	}

	return result
}

// validateReferences 提取并校验文本中的所有 URL
func (s *Service) validateReferences(text string) *ValidationResult {
	result := &ValidationResult{Valid: true}

	matches := urlRe.FindAllString(text, -1)

	if len(matches) > s.urlValidator.cfg.MaxReferenceCount {
		result.AddWarning("引用数量 %d 超过建议值 %d，可能影响性能", len(matches), s.urlValidator.cfg.MaxReferenceCount)
	}

	for _, rawURL := range matches {
		// 去掉末尾标点
		rawURL = strings.TrimRight(rawURL, ".,;:!?，。；：！？")
		if err := s.urlValidator.ValidateURL(rawURL); err != nil {
			result.AddError("引用 URL 不安全: %s — %v", rawURL, err)
		}
	}

	return result
}
