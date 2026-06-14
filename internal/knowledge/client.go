package knowledge

import (
	"fmt"
	"net/http"

	"cogniforge/internal/httpclient"
)

// ServiceClient 是知识库服务调用 Python RAG 服务的客户端。
type ServiceClient struct {
	cli *httpclient.Client
}

func NewServiceClient(cli *httpclient.Client) *ServiceClient {
	return &ServiceClient{cli: cli}
}

// ProcessDocument 调用 Python 服务处理文档并生成向量。
func (c *ServiceClient) ProcessDocument(req *ProcessRequest) (*ProcessResponse, error) {
	var result ProcessResponse
	if err := c.cli.Post(req.Context, "/api/rag/process", req, &result); err != nil {
		return nil, fmt.Errorf("process document: %w", err)
	}
	return &result, nil
}

// Search 调用 Python 服务进行向量检索。
func (c *ServiceClient) Search(req *SearchRequest) (*RAGSearchResponse, error) {
	var result RAGSearchResponse
	if err := c.cli.Post(req.Context, "/api/rag/search", req, &result); err != nil {
		return nil, fmt.Errorf("search: %w", err)
	}
	return &result, nil
}

// DeleteDocument 删除文档的所有向量。
func (c *ServiceClient) DeleteDocument(req *DeleteDocumentRequest) error {
	var result DeleteDocumentResponse
	path := fmt.Sprintf("/api/rag/%s/%s", req.CollectionName, req.DocumentID)
	if err := c.cli.Delete(req.Context, path, nil, &result); err != nil {
		return fmt.Errorf("delete document: %w", err)
	}
	return nil
}

// Health 检查 Python 服务是否可用。
func (c *ServiceClient) Health() bool {
	req, _ := http.NewRequest("GET", c.cli.BaseURL()+"/health", nil)
	resp, err := c.cli.HTTPClient().Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}
