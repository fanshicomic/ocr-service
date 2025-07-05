package main

import (
	"fmt"
	"image"
	"image/jpeg"
	"math"
	"math/rand"
	"net/http"
	"os"
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

// 神奇地将图片裁剪直长宽比为sqrt(π)可以提高识别成功率，不要问我为什么，这就是nyami的玄学力量
func nyamiCrop(img image.Image) image.Image {
	bounds := img.Bounds()
	height := float64(bounds.Dy())
	width := float64(bounds.Dx())

	fmt.Println(height, width, height/width)

	targetRatio := math.Sqrt(math.Pi)
	if height/width < targetRatio {
		newWidth := int(height / targetRatio)
		if newWidth > bounds.Dx() {
			return img
		}

		startX := (bounds.Dx() - newWidth) / 2
		cropRect := image.Rect(startX, 0, startX+newWidth, bounds.Dy())

		if subImg, ok := img.(interface {
			SubImage(r image.Rectangle) image.Image
		}); ok {
			return subImg.SubImage(cropRect)
		}

		cropped := image.NewRGBA(cropRect)
		for y := cropRect.Min.Y; y < cropRect.Max.Y; y++ {
			for x := cropRect.Min.X; x < cropRect.Max.X; x++ {
				cropped.Set(x, y, img.At(x, y))
			}
		}
		return cropped
	}

	return img
}

func saveImageToTmp(img image.Image) string {
	tempFile, _ := os.CreateTemp("", generateRandomString(15))
	defer tempFile.Close()

	jpeg.Encode(tempFile, img, &jpeg.Options{Quality: 100})
	return tempFile.Name()
}

func (s *OCRServer) OCR(c *gin.Context) {
	fileHeader, err := c.FormFile("photo")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "No file is received",
		})
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		return
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return
	}

	croppedImg := nyamiCrop(img)
	tempFilePath := saveImageToTmp(croppedImg)
	fmt.Println(tempFilePath)

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
