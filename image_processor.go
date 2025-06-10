package main

import (
	"fmt"
	"strings"

	"github.com/otiai10/gosseract/v2"
)

type ImageProcessor interface {
	ProcessImage(imagePath string) (map[string]string, error)
}

type ScreenshotImageProcessor struct {
}

func NewScreenshotImageProcessor() ImageProcessor {
	return &ScreenshotImageProcessor{}
}

func (p *ScreenshotImageProcessor) ProcessImage(imagePath string) (map[string]string, error) {
	client := gosseract.NewClient()
	defer client.Close()

	err := client.SetImage(imagePath)
	if err != nil {
		return nil, fmt.Errorf("error setting image: %v", err)
	}

	// **关键**: 设置识别语言为简体中文 (chi_sim)
	// 如果不设置，Tesseract 默认使用英文，无法识别中文
	err = client.SetLanguage("chi_sim")
	if err != nil {
		return nil, fmt.Errorf("error setting language: %v", err)
	}

	// 执行 OCR 并获取纯文本字符串
	ocrResultText, err := client.Text()
	if err != nil {
		return nil, fmt.Errorf("error getting text from image: %v", err)
	}

	fmt.Println("--- Tesseract 原始识别结果 ---")
	fmt.Println(ocrResultText)
	fmt.Println("---------------------------------")

	// ----------------------------------------------------------------
	// 步骤 2: 解析识别出的字符串，提取所需数据
	// ----------------------------------------------------------------

	// 定义我们关心的字段
	targetKeys := map[string]bool{
		"生命": true, "攻击": true, "防御": true,
		"暴击": true, "暴伤": true, "誓约增伤": true,
		"虚弱增伤": true, "誓约回能": true, "加速回能": true,
	}

	// 使用 strings.Fields 按空白符分割字符串
	words := strings.Fields(ocrResultText)

	// 创建一个 map 来存储最终结果
	extractedData := make(map[string]string)

	// 遍历单词切片来匹配 "键" 和 "值"
	for i := 0; i < len(words)-1; i++ {
		// 清理一下识别出的词，有时候会带上奇怪的符号
		cleanWord := strings.Trim(words[i], " ·-:")
		if strings.Contains(cleanWord, "约增伤") {
			cleanWord = "誓约增伤"
		}

		// 检查当前单词是否是我们要找的目标字段
		if _, isTarget := targetKeys[cleanWord]; isTarget {
			// 如果是，那么下一个单词就是它的值
			extractedData[cleanWord] = sanitizeText(words[i+1])
		}
	}

	// ----------------------------------------------------------------
	// 步骤 3: 打印最终提取到的数据
	// ----------------------------------------------------------------
	fmt.Println("\n--- 成功提取的数据 ---")
	if len(extractedData) == 0 {
		fmt.Println("未能提取到任何数据，请检查 Tesseract 原始识别结果是否准确。")
	} else {
		// 为了美观，可以按固定顺序打印
		printOrder := []string{"生命", "攻击", "防御", "暴击", "暴伤", "誓约增伤", "虚弱增伤", "誓约回能", "加速回能"}
		for _, key := range printOrder {
			if value, ok := extractedData[key]; ok {
				fmt.Printf("  %s: %s\n", key, value)
			}
		}
	}

	return extractedData, nil
}

func sanitizeText(text string) string {
	sanizedText := strings.Map(func(r rune) rune {
		if (r >= '0' && r <= '9') || r == '.' {
			return r
		}
		return -1
	}, text)

	return sanizedText
}
