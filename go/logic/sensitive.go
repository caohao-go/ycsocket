// 敏感词过滤模块（对应 PHP sensitive.php + str_replace 过滤逻辑）
package logic

import (
	"bufio"
	"context"
	"os"
	"strings"

	"git.code.oa.com/pcg-csd/trpc-ext/util/log"
)

// SensitiveWords 敏感词列表
var SensitiveWords []string

// InitSensitiveWords 从配置文件加载敏感词（对应 PHP 的 sensitive.php）
// 配置文件格式：每行一个敏感词
func InitSensitiveWords(ctx context.Context) {
	SensitiveWords = make([]string, 0, 16000)

	// 尝试从配置文件加载
	file, err := os.Open("config/sensitive.txt")
	if err != nil {
		log.Warnf(ctx, "load sensitive words file error: %v, sensitive filter disabled", err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		word := strings.TrimSpace(scanner.Text())
		if word != "" {
			SensitiveWords = append(SensitiveWords, word)
		}
	}
	log.Infof(ctx, "loaded %d sensitive words", len(SensitiveWords))
}

// FilterSensitiveWords 过滤敏感词（对应 PHP 的 str_replace($sensitive_words, "*", $content)）
func FilterSensitiveWords(content string) string {
	if len(SensitiveWords) == 0 {
		return content
	}
	for _, word := range SensitiveWords {
		if strings.Contains(content, word) {
			content = strings.ReplaceAll(content, word, "*")
		}
	}
	return content
}

// ContainsSensitiveWords 检查内容是否包含敏感词
// 对应 PHP nicknameSameAction 中: str_replace($sensitive_words, "#", $nickname) + strpos($newnickname, "#") !== false
func ContainsSensitiveWords(content string) bool {
	if len(SensitiveWords) == 0 {
		return false
	}
	for _, word := range SensitiveWords {
		if strings.Contains(content, word) {
			return true
		}
	}
	return false
}
