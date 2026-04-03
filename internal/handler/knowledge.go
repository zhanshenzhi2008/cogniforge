package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func ListKnowledgeBases(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"knowledge_bases": []interface{}{}})
}

func CreateKnowledgeBase(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Create knowledge base"})
}

func GetKnowledgeBase(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"message": "Get knowledge base"}) }

func UpdateKnowledgeBase(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Update knowledge base"})
}

func DeleteKnowledgeBase(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Delete knowledge base"})
}

func UploadDocument(c *gin.Context)  { c.JSON(http.StatusOK, gin.H{"message": "Upload document"}) }
func ListDocuments(c *gin.Context)   { c.JSON(http.StatusOK, gin.H{"documents": []interface{}{}}) }
func DeleteDocument(c *gin.Context)  { c.JSON(http.StatusOK, gin.H{"message": "Delete document"}) }
func SearchKnowledge(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"message": "Search knowledge"}) }
