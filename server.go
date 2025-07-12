package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strings"
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

// 神奇地将图片裁剪直长宽比为 1.6 可以提高识别成功率，不要问我为什么，这就是nyami的玄学力量
func nyamiCrop(img image.Image) image.Image {
	bounds := img.Bounds()
	height := float64(bounds.Dy())

	newHeight := int(height / 2)
	newWidth := int(height / 2 * 0.8)

	startX := (bounds.Dx() - newWidth) / 2
	startY := (bounds.Dy() - int(newHeight)) / 2
	cropRect := image.Rect(startX, startY, startX+newWidth, startY+newHeight)

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

// 转黑白增强对比度
func nyamiBoost(img image.Image) *image.Gray {
	bounds := img.Bounds()
	grayImg := image.NewGray(bounds)
	draw.Draw(grayImg, grayImg.Bounds(), img, bounds.Min, draw.Src)

	// 190的对比度刚刚好
	threshold := uint8(190)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c := grayImg.GrayAt(x, y)
			if c.Y > threshold {
				grayImg.SetGray(x, y, color.Gray{Y: 255})
			} else {
				grayImg.SetGray(x, y, color.Gray{Y: 0})
			}
		}
	}
	return grayImg
}

func saveImageToTmp(img image.Image, format string) string {
	tempFile, _ := os.CreateTemp("", generateRandomString(15))
	defer tempFile.Close()

	switch strings.ToLower(format) {
	case "jpeg", "jpg":
		jpeg.Encode(tempFile, img, &jpeg.Options{Quality: 100})
	case "png":
		png.Encode(tempFile, img)
	default:
		return ""
	}

	return tempFile.Name()
}

func (s *OCRServer) OCR(c *gin.Context) {
	c.Header("Content-Type", "application/json; charset=utf-8")
	fileHeader, err := c.FormFile("photo")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":  "No file is received",
			"detail": err.Error(),
		})
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open image", "detail": err.Error()})
		return
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode image", "detail": err.Error()})
		return
	}

	extension := filepath.Ext(fileHeader.Filename)
	croppedImg := nyamiCrop(img)
	bwImg := nyamiBoost(croppedImg)
	tempFilePath := saveImageToTmp(bwImg, extension[1:])
	fmt.Println(tempFilePath)

	result, err := s.ImageProcessor.ProcessImage(tempFilePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process image", "detail": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"result": result})
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
