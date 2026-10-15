package prompt

type CreatePromptRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Content     string `json:"content"` // 初版正文，可空
	Changelog   string `json:"changelog"`
}

type UpdatePromptRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

type CreateVersionRequest struct {
	Content   string `json:"content" binding:"required"`
	Changelog string `json:"changelog"`
}

type PromptDetail struct {
	Prompt         interface{} `json:"prompt"`
	CurrentVersion interface{} `json:"current_version,omitempty"`
}
