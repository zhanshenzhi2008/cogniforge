package chat

import "strings"

// pythonServiceURL is set from config.RAG.PythonServiceURL.
var pythonServiceURL string

// SetPythonURL sets the Python service URL (called from router setup).
func SetPythonURL(url string) { pythonServiceURL = strings.TrimRight(url, "/") }

// PythonURL 供 app 层 handler 调用记忆摘要/检索。
func PythonURL() string { return pythonServiceURL }
