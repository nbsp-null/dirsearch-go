package dictionary

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"runtime/debug"
	"strings"

	"dirsearch-go/internal/config"
	"dirsearch-go/internal/utils"
)

// Dictionary 字典结构
type Dictionary struct {
	config        *config.Config
	wordlists     []string
	extensions    []string
	prefixes      []string
	suffixes      []string
	words         []string
	sourceFactory *SourceFactory
}

// NewDictionary 创建新的字典
func NewDictionary(cfg *config.Config) (*Dictionary, error) {
	dict := &Dictionary{
		config:        cfg,
		wordlists:     cfg.Dictionary.Wordlists,
		extensions:    cfg.Dictionary.DefaultExtensions,
		prefixes:      cfg.Dictionary.Prefixes,
		suffixes:      cfg.Dictionary.Suffixes,
		words:         make([]string, 0),
		sourceFactory: NewSourceFactory(),
	}

	// 加载字典文件
	if err := dict.loadWordlists(); err != nil {
		return nil, fmt.Errorf("failed to load wordlists: %w", err)
	}

	return dict, nil
}

// loadWordlists 加载字典文件
func (dict *Dictionary) loadWordlists() error {
	for _, wordlistPath := range dict.wordlists {
		// 检查是否为URL，如果是URL则跳过文件加载
		if utils.IsURL(wordlistPath) {
			log.Printf("Debug: Skipping URL wordlist in file loading: %s", wordlistPath)
			continue
		}

		// 检查是否为目录
		if info, err := os.Stat(wordlistPath); err == nil && info.IsDir() {
			// 如果是目录，加载目录下的所有文件
			files, err := filepath.Glob(filepath.Join(wordlistPath, "*"))
			if err != nil {
				return fmt.Errorf("failed to glob directory %s: %w", wordlistPath, err)
			}
			for _, file := range files {
				if err := dict.loadWordlistFile(file); err != nil {
					return fmt.Errorf("failed to load wordlist file %s: %w", file, err)
				}
			}
		} else {
			// 如果是文件，直接加载
			if err := dict.loadWordlistFile(wordlistPath); err != nil {
				return fmt.Errorf("failed to load wordlist file %s: %w", wordlistPath, err)
			}
		}
	}

	// 尝试从配置的源加载wordlist
	if err := dict.loadFromSources(); err != nil {
		return fmt.Errorf("failed to load from sources: %w", err)
	}

	return nil
}

// loadWordlistFile 加载单个字典文件
func (dict *Dictionary) loadWordlistFile(filepath string) error {
	file, err := os.Open(filepath)
	if err != nil {
		return fmt.Errorf("failed to open wordlist file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		word := strings.TrimSpace(scanner.Text())

		// 跳过空行和注释
		if word == "" || strings.HasPrefix(word, "#") {
			continue
		}

		// 应用大小写转换
		if dict.config.Dictionary.Lowercase {
			word = strings.ToLower(word)
		} else if dict.config.Dictionary.Uppercase {
			word = strings.ToUpper(word)
		} else if dict.config.Dictionary.Capitalization {
			word = strings.Title(strings.ToLower(word))
		}

		dict.words = append(dict.words, word)
	}

	return scanner.Err()
}

// loadFromSources 从配置的源加载wordlist
func (dict *Dictionary) loadFromSources() error {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("loadFromSources panic recovered: %v\nStack trace: %s", r, debug.Stack())
		}
	}()

	// 检查是否有配置的源
	if dict.config.Dictionary.Source.Type == "" {
		log.Printf("Debug: No source type configured")
		return nil // 没有配置源，跳过
	}

	log.Printf("Debug: Loading from source type: %s", dict.config.Dictionary.Source.Type)

	// 如果是file类型但没有指定路径，跳过
	if dict.config.Dictionary.Source.Type == "file" && dict.config.Dictionary.Source.Path == "" {
		log.Printf("Debug: File source type but no path specified")
		return nil // file类型但没有路径，跳过
	}

	// 创建源配置
	sourceConfig := &SourceConfig{
		Type:     SourceType(dict.config.Dictionary.Source.Type),
		Path:     dict.config.Dictionary.Source.Path,
		URL:      dict.config.Dictionary.Source.URL,
		DBHost:   dict.config.Dictionary.Source.DBHost,
		DBPort:   dict.config.Dictionary.Source.DBPort,
		DBUser:   dict.config.Dictionary.Source.DBUser,
		DBPass:   dict.config.Dictionary.Source.DBPass,
		DBName:   dict.config.Dictionary.Source.DBName,
		DBTable:  dict.config.Dictionary.Source.DBTable,
		DBColumn: dict.config.Dictionary.Source.DBColumn,
	}

	log.Printf("Debug: Source config - Type: %s, URL: %s", sourceConfig.Type, sourceConfig.URL)

	// 创建源
	source, err := dict.sourceFactory.CreateSource(sourceConfig)
	if err != nil {
		return fmt.Errorf("failed to create source: %w", err)
	}
	defer source.Close()

	// 获取单词
	words, err := source.GetWords()
	if err != nil {
		return fmt.Errorf("failed to get words from source: %w", err)
	}

	log.Printf("Debug: Loaded %d words from source", len(words))

	// 处理单词
	for _, word := range words {
		word = strings.TrimSpace(word)
		if word == "" || strings.HasPrefix(word, "#") {
			continue
		}

		// 应用大小写转换
		if dict.config.Dictionary.Lowercase {
			word = strings.ToLower(word)
		} else if dict.config.Dictionary.Uppercase {
			word = strings.ToUpper(word)
		} else if dict.config.Dictionary.Capitalization {
			word = strings.Title(strings.ToLower(word))
		}

		dict.words = append(dict.words, word)
	}

	log.Printf("Debug: Total words after processing: %d", len(dict.words))

	return nil
}

// GeneratePaths 生成扫描路径
func (dict *Dictionary) GeneratePaths() ([]string, error) {
	var paths []string

	for _, word := range dict.words {
		// 跳过被排除的扩展名
		if dict.shouldExcludeWord(word) {
			continue
		}

		// 处理扩展名
		if dict.config.Dictionary.ForceExtensions {
			// 强制添加扩展名
			paths = append(paths, word)
			for _, ext := range dict.extensions {
				paths = append(paths, word+"."+ext)
			}
			paths = append(paths, word+"/")
		} else if dict.config.Dictionary.OverwriteExtensions {
			// 覆盖扩展名
			paths = append(paths, word)
			for _, ext := range dict.extensions {
				paths = append(paths, dict.replaceExtension(word, ext))
			}
		} else {
			// 替换 %EXT% 关键字
			if strings.Contains(word, "%EXT%") {
				for _, ext := range dict.extensions {
					newWord := strings.ReplaceAll(word, "%EXT%", ext)
					paths = append(paths, newWord)
				}
			} else {
				paths = append(paths, word)
			}
		}

		// 添加前缀
		for _, prefix := range dict.prefixes {
			paths = append(paths, prefix+word)
		}

		// 添加后缀
		for _, suffix := range dict.suffixes {
			// 跳过目录的后缀
			if !strings.HasSuffix(word, "/") {
				paths = append(paths, word+suffix)
			}
		}
	}

	// 去重
	paths = dict.deduplicate(paths)

	return paths, nil
}

// shouldExcludeWord 判断是否应该排除单词
func (dict *Dictionary) shouldExcludeWord(word string) bool {
	for _, excludeExt := range dict.config.Dictionary.ExcludeExtensions {
		if strings.HasSuffix(word, "."+excludeExt) {
			return true
		}
	}
	return false
}

// replaceExtension 替换扩展名
func (dict *Dictionary) replaceExtension(word, newExt string) string {
	// 定义不应该被覆盖的扩展名
	protectedExts := []string{"log", "json", "xml", "jpg", "jpeg", "png", "gif", "bmp", "ico", "svg", "css", "js", "woff", "woff2", "ttf", "eot"}

	// 检查当前扩展名是否受保护
	for _, protectedExt := range protectedExts {
		if strings.HasSuffix(strings.ToLower(word), "."+protectedExt) {
			return word // 返回原单词，不覆盖
		}
	}

	// 替换扩展名
	extRegex := regexp.MustCompile(`\.[a-zA-Z0-9]+$`)
	if extRegex.MatchString(word) {
		return extRegex.ReplaceAllString(word, "."+newExt)
	}

	return word + "." + newExt
}

// deduplicate 去重
func (dict *Dictionary) deduplicate(paths []string) []string {
	seen := make(map[string]bool)
	var result []string

	for _, path := range paths {
		if !seen[path] {
			seen[path] = true
			result = append(result, path)
		}
	}

	return result
}

// GetWordCount 获取单词数量
func (dict *Dictionary) GetWordCount() int {
	return len(dict.words)
}

// GetPathCount 获取路径数量
func (dict *Dictionary) GetPathCount() int {
	paths, _ := dict.GeneratePaths()
	return len(paths)
}
