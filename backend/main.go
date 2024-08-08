package main

import (
	"backend/analyzer"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"net/http"
)

type ExecuteRequest struct {
	Content string `json:"content" binding:"required"`
}

type ExecuteResponse struct {
	Result string `json:"result"`
}

func main() {

	r := gin.Default()

	r.Use(cors.Default())

	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Hello, World!",
		})
	})

	r.POST("/execute", func(c *gin.Context) {
		var req ExecuteRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		result := processContent(req.Content)

		c.JSON(http.StatusOK, ExecuteResponse{
			Result: result,
		})
	})

	err := r.Run(":5000")
	if err != nil {
		return
	}
}

func processContent(content string) string {
	results := analyzer.Analyzer(content)
	return results
}
