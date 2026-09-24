package model

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// =============================================================================
// Skill Models - 技能模板（阶段十五 15.3）
//
// SKILL 结构参照 Claude Skills / GPTs，分三层：
//   meta         - 元数据（名称/描述/版本/标签/作者）
//   instructions - 详细指令，含约束规则和 Few-shot 示例
//   references   - 参考数据（URL/文件/内嵌文本）
//
// Skill 表用 JSONB 存 flexible 部分，保持 schema 简单。
// =============================================================================

// Skill 技能模板表
type Skill struct {
	ID           string         `gorm:"primaryKey;type:varchar(64)" json:"id"`
	UserID       string         `gorm:"type:varchar(64);index" json:"user_id"` // 空 = 内置，全局可用
	Name         string         `gorm:"type:varchar(255);not null" json:"name"`
	Description  string         `gorm:"type:text" json:"description"`
	Icon         string         `gorm:"type:varchar(50)" json:"icon"` // emoji

	// 指令部分（代替原来的 SystemPrompt）
	Instructions string     `gorm:"type:text" json:"instructions"` // 核心指令，含约束和 Few-shot
	Examples     JSONBArray `gorm:"type:jsonb" json:"examples"`     // ["{\"role\":\"user\",\"content\":\"...\"}",...]

	// 参考数据
	References JSONBArray `gorm:"type:jsonb" json:"references"` // [ReferenceItem, ...]

	// 约束规则
	Constraints JSONBArray `gorm:"type:jsonb" json:"constraints"` // ["禁止xxx", "必须xxx", ...]

	// 配置
	Model       string     `gorm:"type:varchar(100)" json:"model"`
	McpServers  JSONBArray `gorm:"type:jsonb" json:"mcp_servers"`  // MCP Server ID 列表
	MemoryType  string     `gorm:"type:varchar(50)" json:"memory_type"`
	MemoryTurns int        `gorm:"default:10" json:"memory_turns"`

	// 分类与标签
	Category string     `gorm:"type:varchar(100)" json:"category"`     // coding/data/research/creative
	Tags     JSONBArray `gorm:"type:jsonb" json:"tags"`               // ["python","sql"]
	Version  string     `gorm:"type:varchar(20)" json:"version"`      // v1.0
	Author   string     `gorm:"type:varchar(100)" json:"author"`      // 作者/来源

	// 元数据
	IsBuiltIn  bool      `gorm:"default:false" json:"is_built_in"`
	SortOrder  int       `gorm:"default:0" json:"sort_order"`
	UsageCount int        `gorm:"default:0" json:"usage_count"` // 被使用次数
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Skill) TableName() string {
	return "skills"
}

// ReferenceItem 参考数据条目
type ReferenceItem struct {
	Type    string `json:"type"`    // "url" | "file" | "text"
	Title   string `json:"title"`   // 显示标题
	Content string `json:"content"`  // 内嵌文本（type=text 时）
	URL     string `json:"url"`     // type=url 时
	Path    string `json:"path"`    // type=file 时
}

// BuildSystemPrompt 把 Skill 拼成完整的 system prompt
// 顺序：基本描述 → constraints → instructions → examples → references
func (s *Skill) BuildSystemPrompt() string {
	var out string

	// 1. 基本角色描述
	if s.Description != "" {
		out += "## 角色描述\n" + s.Description + "\n\n"
	}

	// 2. 约束规则
	if len(s.Constraints) > 0 {
		out += "## 约束规则\n"
		for _, c := range s.Constraints {
			out += "- " + c + "\n"
		}
		out += "\n"
	}

	// 3. 核心指令
	if s.Instructions != "" {
		out += "## 指令\n" + s.Instructions + "\n\n"
	}

	// 4. Few-shot 示例
	if len(s.Examples) > 0 {
		out += "## 示例\n"
		for _, ex := range s.Examples {
			out += ex + "\n"
		}
		out += "\n"
	}

	// 5. 参考数据
	if len(s.References) > 0 {
		out += "## 参考数据\n"
		for _, refRaw := range s.References {
			var ref ReferenceItem
			if err := json.Unmarshal([]byte(refRaw), &ref); err == nil {
				switch ref.Type {
				case "url":
					out += "- [" + ref.Title + "](" + ref.URL + ")\n"
				case "text":
					out += "- **" + ref.Title + "**: " + ref.Content + "\n"
				default:
					out += "- " + refRaw + "\n"
				}
			} else {
				out += "- " + refRaw + "\n"
			}
		}
		out += "\n"
	}

	return out
}

// =============================================================================
// 内置 SKILL 种子数据
// =============================================================================

var BuiltInSkills = []Skill{
	{
		ID:           "code_assistant",
		Name:         "代码助手",
		Description:  "你是一个资深全栈工程师，精通 Python、Go、JavaScript/TypeScript、Vue、React、SQL 等主流语言和框架。",
		Icon:         "💻",
		Instructions: `请遵循以下工作方式：
1. 代码优先：回复时优先给出可运行的代码，注释清晰
2. 解释简洁：不堆砌概念，直接说明「做什么」「怎么做」
3. 错误排查：给出可能的根因和解决方向，附关键日志/配置位置
4. 架构建议：涉及多模块时，给出目录结构或模块边界建议`,
		Examples: JSONBArray{
			`User: 帮我写一个 Go HTTP 中间件记录请求耗时
Assistant: 以下是完整实现...`,
		},
		Constraints: JSONBArray{
			"禁止直接给出未经验证的代码",
			"SQL 操作必须防注入",
			"敏感信息（密码/Key）不要出现在回复里",
		},
		References: JSONBArray{
			`{"type":"url","title":"Go 标准库文档","url":"https://pkg.go.dev"}`,
		},
		Model:      "",
		McpServers: JSONBArray{"code_execute"},
		MemoryType: "short_term",
		MemoryTurns: 10,
		Category:   "coding",
		Tags:       JSONBArray{"go", "python", "javascript", "sql"},
		Version:    "v1.0",
		IsBuiltIn:  true,
		SortOrder:  1,
	},
	{
		ID:           "web_assistant",
		Name:         "网页助手",
		Description:  "你是一个专业信息收集整理助手，擅长通过网页搜索和抓取获取最新信息，能对多个来源进行综合整理，给出客观、结构化的回复。",
		Icon:         "🌐",
		Instructions: `请遵循以下工作方式：
1. 先搜索/抓取，再整理：不要只靠训练知识，先查实时信息
2. 引用溯源：每个结论要标注来源（URL 或标题）
3. 结构化输出：用标题/列表/表格组织信息，不要堆砌段落
4. 客观中立：多个来源矛盾时，分别列出，不要替用户选`,
		Examples: JSONBArray{
			`User: 帮我查一下最新的 AI 模型排行榜
Assistant: 根据最新搜索结果...（引用各来源）`,
		},
		Constraints: JSONBArray{
			"必须实际访问 URL 获取内容，不能虚构引用",
			"无法访问的页面要明确告知用户",
			"禁止替用户做涉及金钱/法律的决策",
		},
		References: JSONBArray{
			`{"type":"url","title":"Bing 搜索","url":"https://www.bing.com"}`,
		},
		Model:      "",
		McpServers: JSONBArray{"fetch", "web_search"},
		MemoryType: "short_term",
		MemoryTurns: 10,
		Category:   "research",
		Tags:       JSONBArray{"search", "research", "report"},
		Version:    "v1.0",
		IsBuiltIn:  true,
		SortOrder:  2,
	},
	{
		ID:           "data_analyst",
		Name:         "数据分析",
		Description:  "你是一个专业数据分析师，擅长 SQL 查询、数据清洗、统计分析、数据可视化。",
		Icon:         "📊",
		Instructions: `请遵循以下工作方式：
1. 结论先行：先给核心发现，再说数据细节
2. 数据说话：所有结论要有具体数字支撑，注明单位
3. 可视化友好：图表建议要给出图表类型和关键维度
4. SQL 规范：JOIN 必须注明关联条件，避免笛卡尔积`,
		Examples: JSONBArray{
			`User: 分析一下这个月用户活跃趋势
Assistant: 核心发现：DAU 环比增长 12%，关键变化在...`,
		},
		Constraints: JSONBArray{
			"数据量估算要说明假设前提",
			"图表代码必须可运行",
			"涉及用户隐私的数据不能直接展示",
		},
		Model:      "",
		McpServers: JSONBArray{"code_execute"},
		MemoryType: "short_term",
		MemoryTurns: 10,
		Category:   "data",
		Tags:       JSONBArray{"sql", "python", "pandas", "visualization"},
		Version:    "v1.0",
		IsBuiltIn:  true,
		SortOrder:  3,
	},
}

// GetBuiltInSkill 根据 ID 查找内置 SKILL
func GetBuiltInSkill(id string) *Skill {
	for i := range BuiltInSkills {
		if BuiltInSkills[i].ID == id {
			return &BuiltInSkills[i]
		}
	}
	return nil
}

// IsBuiltInSkill 判断是否为内置 SKILL
func IsBuiltInSkill(id string) bool {
	return GetBuiltInSkill(id) != nil
}
