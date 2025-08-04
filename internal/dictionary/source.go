package dictionary

import (
	"bufio"
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// SourceType wordlist源类型
type SourceType string

const (
	SourceFile SourceType = "file"
	SourceURL  SourceType = "url"
	SourceDB   SourceType = "database"
)

// SourceConfig wordlist源配置
type SourceConfig struct {
	Type     SourceType `mapstructure:"type"`
	Path     string     `mapstructure:"path"`
	URL      string     `mapstructure:"url"`
	DBHost   string     `mapstructure:"db-host"`
	DBPort   int        `mapstructure:"db-port"`
	DBUser   string     `mapstructure:"db-user"`
	DBPass   string     `mapstructure:"db-password"`
	DBName   string     `mapstructure:"db-name"`
	DBTable  string     `mapstructure:"db-table"`
	DBColumn string     `mapstructure:"db-column"`
}

// WordlistSource wordlist源接口
type WordlistSource interface {
	GetWords() ([]string, error)
	Close() error
}

// FileSource 文件源
type FileSource struct {
	path string
	file *os.File
}

// NewFileSource 创建文件源
func NewFileSource(path string) *FileSource {
	return &FileSource{path: path}
}

// GetWords 从文件获取单词
func (fs *FileSource) GetWords() ([]string, error) {
	file, err := os.Open(fs.path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", fs.path, err)
	}
	fs.file = file
	defer file.Close()

	var words []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		word := strings.TrimSpace(scanner.Text())
		if word != "" && !strings.HasPrefix(word, "#") {
			words = append(words, word)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	return words, nil
}

// Close 关闭文件源
func (fs *FileSource) Close() error {
	if fs.file != nil {
		return fs.file.Close()
	}
	return nil
}

// URLSource URL源
type URLSource struct {
	url      string
	client   *http.Client
	response *http.Response
}

// NewURLSource 创建URL源
func NewURLSource(url string) *URLSource {
	return &URLSource{
		url: url,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetWords 从URL获取单词
func (us *URLSource) GetWords() ([]string, error) {
	resp, err := us.client.Get(us.url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch URL %s: %w", us.url, err)
	}
	us.response = resp
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	lines := strings.Split(string(body), "\n")
	var words []string
	for _, line := range lines {
		word := strings.TrimSpace(line)
		if word != "" && !strings.HasPrefix(word, "#") {
			words = append(words, word)
		}
	}

	return words, nil
}

// Close 关闭URL源
func (us *URLSource) Close() error {
	if us.response != nil {
		return us.response.Body.Close()
	}
	return nil
}

// DBSource 数据库源
type DBSource struct {
	config *SourceConfig
	db     *sql.DB
}

// NewDBSource 创建数据库源
func NewDBSource(config *SourceConfig) *DBSource {
	return &DBSource{config: config}
}

// GetWords 从数据库获取单词
func (ds *DBSource) GetWords() ([]string, error) {
	// 构建DSN
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		ds.config.DBUser,
		ds.config.DBPass,
		ds.config.DBHost,
		ds.config.DBPort,
		ds.config.DBName,
	)

	// 连接数据库
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	ds.db = db

	// 测试连接
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// 查询单词
	query := fmt.Sprintf("SELECT %s FROM %s", ds.config.DBColumn, ds.config.DBTable)
	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query database: %w", err)
	}
	defer rows.Close()

	var words []string
	for rows.Next() {
		var word string
		if err := rows.Scan(&word); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		word = strings.TrimSpace(word)
		if word != "" {
			words = append(words, word)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return words, nil
}

// Close 关闭数据库源
func (ds *DBSource) Close() error {
	if ds.db != nil {
		return ds.db.Close()
	}
	return nil
}

// SourceFactory 源工厂
type SourceFactory struct{}

// NewSourceFactory 创建源工厂
func NewSourceFactory() *SourceFactory {
	return &SourceFactory{}
}

// CreateSource 创建wordlist源
func (sf *SourceFactory) CreateSource(config *SourceConfig) (WordlistSource, error) {
	switch config.Type {
	case SourceFile:
		return NewFileSource(config.Path), nil
	case SourceURL:
		return NewURLSource(config.URL), nil
	case SourceDB:
		return NewDBSource(config), nil
	default:
		return nil, fmt.Errorf("unsupported source type: %s", config.Type)
	}
}
