package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
)

type Server interface {
	Ping(c *gin.Context)
	ProcessImage(c *gin.Context)
	OCR(c *gin.Context)
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
	c.JSON(http.StatusOK, gin.H{
		"status":  "OK",
		"message": "OCR Server is running",
	})
}

func (s *OCRServer) ProcessImage(c *gin.Context) {
	imagePath := c.PostForm("image_path")
	if imagePath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "image_path is required"})
		return
	}

	result, err := s.ImageProcessor.ProcessImage(imagePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process image", "detail": err.Error()})
		return
	}

	c.JSON(200, gin.H{"result": result})
}

func (s *OCRServer) OCR(c *gin.Context) {
	file, err := c.FormFile("photo")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "No file is received",
		})
		return
	}

	extension := filepath.Ext(file.Filename)
	tempFilePath := filepath.Join(os.TempDir(), fmt.Sprintf("%s%s", generateRandomString(15), extension))

	if err := c.SaveUploadedFile(file, tempFilePath); err != nil {
		log.Printf("Failed to save file: %s", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Unable to save the file",
		})
		return
	}

	result, err := s.ImageProcessor.ProcessImage(tempFilePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process image", "detail": err.Error()})
		return
	}

	c.JSON(200, gin.H{"result": result})
}

func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	var seededRand *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))

	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}
