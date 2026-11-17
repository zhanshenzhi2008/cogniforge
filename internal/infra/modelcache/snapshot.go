package modelcache

// Redis / 本地共用的模型配置快照。
// encrypted_key 仍是库里的密文，明文 Key 只存在 Go 进程内存，不进 Redis。

const (
	KeyRev      = "cogniforge:modelcfg:rev"
	KeySnapshot = "cogniforge:modelcfg:snapshot"
	RedisTTLSec = 600
)

type ModelItem struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Snapshot struct {
	Rev          int64             `json:"rev"`
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Provider     string            `json:"provider"`
	BaseURL      string            `json:"base_url"`
	DefaultModel string            `json:"default_model"`
	ExtraHeaders map[string]string `json:"extra_headers,omitempty"`
	EncryptedKey string            `json:"encrypted_key,omitempty"`
	Models       []ModelItem       `json:"models,omitempty"`
}
