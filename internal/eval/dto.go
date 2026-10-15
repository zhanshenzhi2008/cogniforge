package eval

type CreateSuiteRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

type UpdateSuiteRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

type CreateCaseRequest struct {
	Name        string                 `json:"name"`
	Input       string                 `json:"input" binding:"required"`
	Expected    string                 `json:"expected"`
	JudgeType   string                 `json:"judge_type"`
	JudgeConfig map[string]interface{} `json:"judge_config"`
	Weight      *float64               `json:"weight"`
	SortOrder   *int                   `json:"sort_order"`
}

type UpdateCaseRequest struct {
	Name        *string                `json:"name"`
	Input       *string                `json:"input"`
	Expected    *string                `json:"expected"`
	JudgeType   *string                `json:"judge_type"`
	JudgeConfig map[string]interface{} `json:"judge_config"`
	Weight      *float64               `json:"weight"`
	SortOrder   *int                   `json:"sort_order"`
}

type CreateRunRequest struct {
	TargetType string `json:"target_type" binding:"required"` // model | agent
	TargetID   string `json:"target_id" binding:"required"`
}

type RunDetailResponse struct {
	Run     interface{} `json:"run"`
	Results interface{} `json:"results"`
}
