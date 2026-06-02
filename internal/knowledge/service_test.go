package knowledge

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"cogniforge/internal/database"
	"cogniforge/internal/model"
)

// ============ DTO Tests ============

func TestCreateKBRequest_JSON(t *testing.T) {
	jsonData := `{
		"name": "Test KB",
		"description": "A test",
		"vector_db": "chroma",
		"embedding_model": "text-embedding-ada-002"
	}`

	var req CreateKBRequest
	err := json.Unmarshal([]byte(jsonData), &req)

	assert.NoError(t, err)
	assert.Equal(t, "Test KB", req.Name)
	assert.Equal(t, "A test", req.Description)
	assert.Equal(t, "chroma", req.VectorDB)
	assert.Equal(t, "text-embedding-ada-002", req.EmbeddingModel)
}

func TestUpdateKBRequest_JSON(t *testing.T) {
	jsonData := `{
		"name": "Updated Name",
		"status": "disabled"
	}`

	var req UpdateKBRequest
	err := json.Unmarshal([]byte(jsonData), &req)

	assert.NoError(t, err)
	assert.Equal(t, "Updated Name", req.Name)
	assert.Equal(t, "disabled", req.Status)
	assert.Empty(t, req.Description)
}

func TestSearchRequest_JSON(t *testing.T) {
	jsonData := `{
		"query": "machine learning",
		"top_k": 10,
		"min_score": 0.5
	}`

	var req SearchRequest
	err := json.Unmarshal([]byte(jsonData), &req)

	assert.NoError(t, err)
	assert.Equal(t, "machine learning", req.Query)
	assert.Equal(t, 10, req.TopK)
	assert.Equal(t, 0.5, req.MinScore)
}

func TestSearchResponse_JSON(t *testing.T) {
	response := SearchResponse{
		Results: []SearchResult{
			{
				DocumentID:   "doc_123",
				DocumentName: "test.pdf",
				ChunkID:      "doc_123_chunk_0",
				Content:      "This is the chunk content...",
				Score:        0.85,
			},
		},
		Total:    1,
		Query:    "test query",
		Duration: 150,
	}

	data, err := json.Marshal(response)
	assert.NoError(t, err)

	var decoded SearchResponse
	err = json.Unmarshal(data, &decoded)
	assert.NoError(t, err)
	assert.Len(t, decoded.Results, 1)
	assert.Equal(t, 0.85, decoded.Results[0].Score)
}

// ============ Helper Function Tests ============

func TestGenerateDocID(t *testing.T) {
	id1 := generateDocID()
	id2 := generateDocID()

	assert.NotEmpty(t, id1)
	assert.NotEmpty(t, id2)
	assert.NotEqual(t, id1, id2)
	assert.Contains(t, id1, "doc_")
}

func TestGetFileExt(t *testing.T) {
	tests := []struct {
		filename string
		expected string
	}{
		{"document.pdf", ".pdf"},
		{"DOCUMENT.PDF", ".pdf"},
		{"document.txt", ".txt"},
		{"document.md", ".md"},
		{"document.docx", ".docx"},
		{"document.html", ".html"},
		{"noextension", ""},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			result := getFileExt(tt.filename)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetFileType(t *testing.T) {
	tests := []struct {
		filename string
		expected string
		ok       bool
	}{
		{"doc.pdf", "pdf", true},
		{"doc.txt", "txt", true},
		{"doc.md", "md", true},
		{"doc.docx", "docx", true},
		{"doc.html", "html", true},
		{"doc.exe", "", false},
		{"doc", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			fileType, ok := getFileType(tt.filename)
			assert.Equal(t, tt.expected, fileType)
			assert.Equal(t, tt.ok, ok)
		})
	}
}

func TestChunkText(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		chunkSize int
		overlap   int
		minChunks int
	}{
		{
			name:      "short content",
			content:   "Hello world",
			chunkSize: 512,
			overlap:   50,
			minChunks: 1,
		},
		{
			name:      "medium content",
			content:   "This is a test document. " + string(make([]byte, 200)),
			chunkSize: 50,
			overlap:   10,
			minChunks: 1,
		},
		{
			name:      "empty content",
			content:   "",
			chunkSize: 512,
			overlap:   50,
			minChunks: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chunks := chunkText(tt.content, tt.chunkSize, tt.overlap)
			assert.GreaterOrEqual(t, len(chunks), tt.minChunks)
		})
	}
}

func TestTruncateText(t *testing.T) {
	tests := []struct {
		input    string
		maxLen   int
		expected string
	}{
		{"short", 10, "short"},
		{"This is a long text", 10, "This is a ..."},
		{"exact", 5, "exact"},
		{"", 10, ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := truncateText(tt.input, tt.maxLen)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTokenize(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"hello world", 2},
		{"hello, world!", 2},
		{"  spaces  multiple   spaces  ", 2},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			tokens := tokenize(tt.input)
			assert.GreaterOrEqual(t, len(tokens), tt.expected)
		})
	}
}

func TestStripHTMLTags(t *testing.T) {
	html := `<html><body><h1>Title</h1><p>Paragraph with <strong>bold</strong> text.</p></body></html>`
	result := stripHTMLTags(html)

	assert.NotContains(t, result, "<")
	assert.NotContains(t, result, ">")
	assert.Contains(t, result, "Title")
	assert.Contains(t, result, "Paragraph")
	assert.Contains(t, result, "bold")
}

func TestCalculateTextSimilarity(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		text     string
		minScore float64
	}{
		{
			name:     "exact match",
			query:    "machine learning",
			text:     "machine learning is great",
			minScore: 0.5,
		},
		{
			name:     "partial match",
			query:    "machine",
			text:     "learning machine",
			minScore: 0.3,
		},
		{
			name:     "no match",
			query:    "python",
			text:     "golang is awesome",
			minScore: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			queryWords := tokenize(tt.query)
			score := calculateTextSimilarity(tt.query, tt.text, queryWords)
			assert.GreaterOrEqual(t, score, tt.minScore)
		})
	}
}

// ============ Service Tests (DB Required) ============

func TestKnowledgeService_CRUD(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping DB test in short mode")
	}

	userID := "test_kb_crud"
	database.DB.Where("user_id = ?", userID).Delete(&model.KnowledgeBase{})

	service := NewKnowledgeService(nil)

	// Create
	kb, err := service.CreateKnowledgeBase(userID, &CreateKBRequest{
		Name:           "Test KB",
		Description:    "Test Description",
		VectorDB:       "chroma",
		EmbeddingModel: "text-embedding-ada-002",
	})
	require.NoError(t, err)
	assert.NotEmpty(t, kb.ID)

	// Read
	retrieved, err := service.GetKnowledgeBase(userID, kb.ID)
	require.NoError(t, err)
	assert.Equal(t, "Test KB", retrieved.Name)

	// Update
	updated, err := service.UpdateKnowledgeBase(userID, kb.ID, &UpdateKBRequest{
		Name: "Updated KB",
	})
	require.NoError(t, err)
	assert.Equal(t, "Updated KB", updated.Name)

	// Delete
	err = service.DeleteKnowledgeBase(userID, kb.ID)
	require.NoError(t, err)

	// Verify deleted
	_, err = service.GetKnowledgeBase(userID, kb.ID)
	assert.Error(t, err)

	database.DB.Where("user_id = ?", userID).Delete(&model.KnowledgeBase{})
}

func TestKnowledgeService_List(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping DB test in short mode")
	}

	userID := "test_kb_list"
	database.DB.Where("user_id = ?", userID).Delete(&model.KnowledgeBase{})

	service := NewKnowledgeService(nil)

	// Create multiple KBs
	for i := 0; i < 3; i++ {
		_, err := service.CreateKnowledgeBase(userID, &CreateKBRequest{
			Name: "Test KB",
		})
		require.NoError(t, err)
	}

	// List
	kbs, err := service.ListKnowledgeBases(userID)
	require.NoError(t, err)
	assert.Len(t, kbs, 3)

	// Clean up
	database.DB.Where("user_id = ?", userID).Delete(&model.KnowledgeBase{})
}

func TestKnowledgeService_Defaults(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping DB test in short mode")
	}

	userID := "test_kb_defaults"
	database.DB.Where("user_id = ?", userID).Delete(&model.KnowledgeBase{})

	service := NewKnowledgeService(nil)

	kb, err := service.CreateKnowledgeBase(userID, &CreateKBRequest{
		Name: "Minimal KB",
	})
	require.NoError(t, err)
	assert.Equal(t, "chroma", kb.VectorDB)
	assert.Equal(t, "text-embedding-ada-002", kb.EmbeddingModel)
	assert.Equal(t, "active", kb.Status)

	database.DB.Where("user_id = ?", userID).Delete(&model.KnowledgeBase{})
}

func TestKnowledgeService_Search(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping DB test in short mode")
	}

	userID := "test_kb_search"
	database.DB.Where("user_id = ?", userID).Delete(&model.KnowledgeBase{})

	service := NewKnowledgeService(nil)

	kb, err := service.CreateKnowledgeBase(userID, &CreateKBRequest{
		Name: "Search Test KB",
	})
	require.NoError(t, err)

	// Search empty KB
	result, err := service.SearchKnowledge(userID, kb.ID, &SearchRequest{
		Query: "test query",
	})
	require.NoError(t, err)
	assert.Empty(t, result.Results)
	assert.Equal(t, 0, result.Total)

	database.DB.Where("user_id = ?", userID).Delete(&model.KnowledgeBase{})
}

func TestEnsureUploadDir(t *testing.T) {
	dir, err := ensureUploadDir("user123", "kb456")
	assert.NoError(t, err)
	assert.Contains(t, dir, "uploads")
	assert.Contains(t, dir, "user123")
	assert.Contains(t, dir, "kb456")

	// Clean up
	os.RemoveAll("uploads")
}
