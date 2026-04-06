package handler

import (
	"github.com/gin-gonic/gin"

	"cogniforge/internal/model"
)

func ListKnowledgeBases(c *gin.Context) {
	model.Success(c, gin.H{"knowledge_bases": []interface{}{}})
}

func CreateKnowledgeBase(c *gin.Context) {
	model.Created(c, gin.H{"message": "Create knowledge base"})
}

func GetKnowledgeBase(c *gin.Context) { model.Success(c, gin.H{"message": "Get knowledge base"}) }

func UpdateKnowledgeBase(c *gin.Context) {
	model.Success(c, gin.H{"message": "Update knowledge base"})
}

func DeleteKnowledgeBase(c *gin.Context) {
	model.SuccessWithMessage(c, nil, "Delete knowledge base")
}

func UploadDocument(c *gin.Context) {
	model.Created(c, gin.H{"message": "Upload document"})
}

func ListDocuments(c *gin.Context) {
	model.Success(c, gin.H{"documents": []interface{}{}})
}

func DeleteDocument(c *gin.Context) {
	model.SuccessWithMessage(c, nil, "Delete document")
}

func SearchKnowledge(c *gin.Context) {
	model.Success(c, gin.H{"message": "Search knowledge"})
}
