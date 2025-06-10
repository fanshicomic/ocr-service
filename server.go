package main

import (
	"github.com/gin-gonic/gin"
)

type Server interface {
	Ping(c *gin.Context)
	ProcessImage(c *gin.Context)
}

type OCRServer struct {
	ImageProcessor ImageProcessor
}

func NewOCRServer(imageProcessor ImageProcessor) Server {
	return &OCRServer{
		ImageProcessor: imageProcessor,
	}
}

func (s *OCRServer) Ping(c *gin.Context) {
	c.JSON(200, gin.H{
		"status":  "OK",
		"message": "OCR Server is running",
	})
}

func (s *OCRServer) ProcessImage(c *gin.Context) {
	imagePath := c.PostForm("image_path")
	if imagePath == "" {
		c.JSON(400, gin.H{"error": "image_path is required"})
		return
	}

	result, err := s.ImageProcessor.ProcessImage(imagePath)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to process image", "detail": err.Error()})
		return
	}

	c.JSON(200, gin.H{"result": result})
}
