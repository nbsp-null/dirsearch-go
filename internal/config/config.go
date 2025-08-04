package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

// Config 配置结构
type Config struct {
	General    GeneralConfig    `mapstructure:"general"`
	Dictionary DictionaryConfig `mapstructure:"dictionary"`
	Request    RequestConfig    `mapstructure:"request"`
	Connection ConnectionConfig `mapstructure:"connection"`
	Advanced   AdvancedConfig   `mapstructure:"advanced"`
	View       ViewConfig       `mapstructure:"view"`
	Output     OutputConfig     `mapstructure:"output"`
}

// GeneralConfig 通用配置
type GeneralConfig struct {
	Threads           int      `mapstructure:"threads"`
	Async             bool     `mapstructure:"async"`
	Recursive         bool     `mapstructure:"recursive"`
	DeepRecursive     bool     `mapstructure:"deep-recursive"`
	ForceRecursive    bool     `mapstructure:"force-recursive"`
	RecursionStatus   []string `mapstructure:"recursion-status"`
	MaxRecursionDepth int      `mapstructure:"max-recursion-depth"`
	ExcludeSubdirs    []string `mapstructure:"exclude-subdirs"`
	RandomUserAgents  bool     `mapstructure:"random-user-agents"`
	MaxTime           int      `mapstructure:"max-time"`
	ExitOnError       bool     `mapstructure:"exit-on-error"`
	IncludeStatus     []string `mapstructure:"include-status"`
	ExcludeStatus     []string `mapstructure:"exclude-status"`
	ExcludeSizes      []string `mapstructure:"exclude-sizes"`
	ExcludeText       []string `mapstructure:"exclude-text"`
	ExcludeRegex      []string `mapstructure:"exclude-regex"`
	ExcludeRedirect   []string `mapstructure:"exclude-redirect"`
	ExcludeResponse   []string `mapstructure:"exclude-response"`
	SkipOnStatus      []string `mapstructure:"skip-on-status"`
	MinResponseSize   int      `mapstructure:"min-response-size"`
	MaxResponseSize   int      `mapstructure:"max-response-size"`
}

// DictionaryConfig 字典配置
type DictionaryConfig struct {
	DefaultExtensions   []string     `mapstructure:"default-extensions"`
	ForceExtensions     bool         `mapstructure:"force-extensions"`
	OverwriteExtensions bool         `mapstructure:"overwrite-extensions"`
	Lowercase           bool         `mapstructure:"lowercase"`
	Uppercase           bool         `mapstructure:"uppercase"`
	Capitalization      bool         `mapstructure:"capitalization"`
	ExcludeExtensions   []string     `mapstructure:"exclude-extensions"`
	Prefixes            []string     `mapstructure:"prefixes"`
	Suffixes            []string     `mapstructure:"suffixes"`
	Wordlists           []string     `mapstructure:"wordlists"`
	Source              SourceConfig `mapstructure:"source"`
}

// SourceConfig wordlist源配置
type SourceConfig struct {
	Type     string `mapstructure:"type"`
	Path     string `mapstructure:"path"`
	URL      string `mapstructure:"url"`
	DBHost   string `mapstructure:"db-host"`
	DBPort   int    `mapstructure:"db-port"`
	DBUser   string `mapstructure:"db-user"`
	DBPass   string `mapstructure:"db-password"`
	DBName   string `mapstructure:"db-name"`
	DBTable  string `mapstructure:"db-table"`
	DBColumn string `mapstructure:"db-column"`
}

// RequestConfig 请求配置
type RequestConfig struct {
	HTTPMethod      string   `mapstructure:"http-method"`
	FollowRedirects bool     `mapstructure:"follow-redirects"`
	HeadersFile     string   `mapstructure:"headers-file"`
	UserAgent       string   `mapstructure:"user-agent"`
	Cookie          string   `mapstructure:"cookie"`
	Data            string   `mapstructure:"data"`
	DataFile        string   `mapstructure:"data-file"`
	Headers         []string `mapstructure:"headers"`
	Auth            string   `mapstructure:"auth"`
	AuthType        string   `mapstructure:"auth-type"`
}

// ConnectionConfig 连接配置
type ConnectionConfig struct {
	Timeout             float64  `mapstructure:"timeout"`
	Delay               float64  `mapstructure:"delay"`
	MaxRate             int      `mapstructure:"max-rate"`
	MaxRetries          int      `mapstructure:"max-retries"`
	DomainCheckTimeout  float64  `mapstructure:"domain-check-timeout"`
	DomainCheckRetries  int      `mapstructure:"domain-check-retries"`
	HeadlessTimeout     float64  `mapstructure:"headless-timeout"`
	HeadlessConcurrency int      `mapstructure:"headless-concurrency"`
	Scheme              string   `mapstructure:"scheme"`
	Proxy               string   `mapstructure:"proxy"`
	ProxyFile           string   `mapstructure:"proxy-file"`
	ReplayProxy         string   `mapstructure:"replay-proxy"`
	Proxies             []string `mapstructure:"proxies"`
}

// AdvancedConfig 高级配置
type AdvancedConfig struct {
	Crawl bool `mapstructure:"crawl"`
}

// ViewConfig 视图配置
type ViewConfig struct {
	FullURL              bool `mapstructure:"full-url"`
	QuietMode            bool `mapstructure:"quiet-mode"`
	Color                bool `mapstructure:"color"`
	ShowRedirectsHistory bool `mapstructure:"show-redirects-history"`
	RealTimeStatus       bool `mapstructure:"real-time-status"`
	Headless             bool `mapstructure:"headless"`
	ShowAllStatus        bool `mapstructure:"show-all-status"`
	RecursiveScan        bool `mapstructure:"recursive-scan"`
}

// OutputConfig 输出配置
type OutputConfig struct {
	ReportFormat         string `mapstructure:"report-format"`
	AutosaveReport       bool   `mapstructure:"autosave-report"`
	AutosaveReportFolder string `mapstructure:"autosave-report-folder"`
	LogFile              string `mapstructure:"log-file"`
	LogFileSize          int    `mapstructure:"log-file-size"`
}

var (
	// GlobalConfig 全局配置实例
	GlobalConfig *Config
	// ConfigFile 配置文件路径
	ConfigFile string
)

// Init 初始化配置
func Init() error {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Config Init panic recovered: %v\nStack trace: %s", r, debug.Stack())
		}
	}()

	// 1. 优先加载.env文件（忽略错误）
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: Failed to load .env file: %v", err)
	}

	// 2. 设置环境变量映射
	setupEnvMapping()

	// 3. 设置默认值
	setDefaults()

	// 4. 尝试读取config.ini（如果存在且格式正确）
	configName := "config.ini"
	if envConfig := os.Getenv("DIRSEARCH_CONFIG"); envConfig != "" {
		configName = envConfig
	}

	// 查找配置文件
	configPaths := []string{
		".",
		"./config",
		"./conf",
		os.Getenv("HOME"),
		os.Getenv("USERPROFILE"),
	}

	for _, path := range configPaths {
		if path != "" {
			configFile := filepath.Join(path, configName)
			if _, err := os.Stat(configFile); err == nil {
				// 尝试读取配置文件
				viper.SetConfigFile(configFile)
				viper.SetConfigType("ini")
				if err := viper.ReadInConfig(); err == nil {
					// 成功读取配置文件
					ConfigFile = configFile
					break
				} else {
					log.Printf("Warning: Failed to read config file %s: %v", configFile, err)
				}
			}
		}
	}

	// 5. 如果没有找到有效的配置文件，使用内置默认配置
	if ConfigFile == "" {
		log.Println("No valid config file found, using built-in default configuration")
		// 使用内置默认配置
		if err := viper.ReadConfig(strings.NewReader(defaultConfigINI)); err != nil {
			return fmt.Errorf("failed to load default config: %w", err)
		}
	}

	// 6. 解析配置
	GlobalConfig = &Config{}
	if err := viper.Unmarshal(GlobalConfig); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// 7. 验证配置
	if err := validateConfig(GlobalConfig); err != nil {
		log.Printf("Warning: Config validation failed: %v", err)
	}

	return nil
}

// validateConfig 验证配置
func validateConfig(cfg *Config) error {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("validateConfig panic recovered: %v", r)
		}
	}()

	if cfg == nil {
		return fmt.Errorf("config is nil")
	}

	// 验证线程数
	if cfg.General.Threads <= 0 {
		cfg.General.Threads = 25
	}

	// 验证超时时间
	if cfg.Connection.Timeout <= 0 {
		cfg.Connection.Timeout = 7.5
	}

	// 验证延迟时间
	if cfg.Connection.Delay < 0 {
		cfg.Connection.Delay = 0
	}

	return nil
}

// setupEnvMapping 设置环境变量映射
func setupEnvMapping() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("setupEnvMapping panic recovered: %v", r)
		}
	}()

	viper.SetEnvPrefix("DIRSEARCH")
	viper.AutomaticEnv()

	// 通用配置映射
	viper.BindEnv("general.threads", "DIRSEARCH_THREADS")
	viper.BindEnv("general.max-time", "DIRSEARCH_MAX_TIME")
	viper.BindEnv("general.exit-on-error", "DIRSEARCH_EXIT_ON_ERROR")

	// 字典配置映射
	viper.BindEnv("dictionary.wordlists", "DIRSEARCH_WORDLISTS")
	viper.BindEnv("dictionary.default-extensions", "DIRSEARCH_EXTENSIONS")

	// 连接配置映射
	viper.BindEnv("connection.timeout", "DIRSEARCH_TIMEOUT")
	viper.BindEnv("connection.delay", "DIRSEARCH_DELAY")
	viper.BindEnv("connection.proxy", "DIRSEARCH_PROXY")

	// 请求配置映射
	viper.BindEnv("request.user-agent", "DIRSEARCH_USER_AGENT")
	viper.BindEnv("request.http-method", "DIRSEARCH_HTTP_METHOD")
	viper.BindEnv("request.headers", "DIRSEARCH_HEADERS")

	// 视图配置映射
	viper.BindEnv("view.show-all-status", "DIRSEARCH_SHOW_ALL_STATUS")
	viper.BindEnv("view.recursive-scan", "DIRSEARCH_RECURSIVE_SCAN")
	viper.BindEnv("view.real-time-status", "DIRSEARCH_REAL_TIME_STATUS")
	viper.BindEnv("view.headless", "DIRSEARCH_HEADLESS")

	// 输出配置映射
	viper.BindEnv("output.report-format", "DIRSEARCH_REPORT_FORMAT")
	viper.BindEnv("output.autosave-report", "DIRSEARCH_AUTOSAVE_REPORT")
}

// setDefaults 设置默认值
func setDefaults() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("setDefaults panic recovered: %v", r)
		}
	}()

	// 通用配置默认值
	viper.SetDefault("general.threads", 25)
	viper.SetDefault("general.max-time", 0)
	viper.SetDefault("general.exit-on-error", false)

	// 连接配置默认值
	viper.SetDefault("connection.timeout", 7.5)
	viper.SetDefault("connection.delay", 0)
	viper.SetDefault("connection.max-retries", 3)
	viper.SetDefault("connection.domain-check-timeout", 60)
	viper.SetDefault("connection.domain-check-retries", 3)

	// 请求配置默认值
	viper.SetDefault("request.http-method", "GET")
	viper.SetDefault("request.follow-redirects", false)

	// 视图配置默认值
	viper.SetDefault("view.show-all-status", false)
	viper.SetDefault("view.recursive-scan", false)
	viper.SetDefault("view.real-time-status", false)
	viper.SetDefault("view.headless", false)
	viper.SetDefault("view.color", true)

	// 输出配置默认值
	viper.SetDefault("output.report-format", "plain")
	viper.SetDefault("output.autosave-report", false)
}

// GetConfig 获取配置
func GetConfig() *Config {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("GetConfig panic recovered: %v", r)
		}
	}()

	return GlobalConfig
}

// ParseStatusCodes 解析状态码字符串
func ParseStatusCodes(statusStr string) ([]int, error) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("ParseStatusCodes panic recovered: %v", r)
		}
	}()

	if statusStr == "" {
		return []int{}, nil
	}

	var codes []int
	parts := strings.Split(statusStr, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// 处理范围，如 "200-299"
		if strings.Contains(part, "-") {
			rangeParts := strings.Split(part, "-")
			if len(rangeParts) == 2 {
				start := strings.TrimSpace(rangeParts[0])
				end := strings.TrimSpace(rangeParts[1])

				startCode, err := parseInt(start)
				if err != nil {
					continue
				}

				endCode, err := parseInt(end)
				if err != nil {
					continue
				}

				for i := startCode; i <= endCode; i++ {
					codes = append(codes, i)
				}
			}
		} else {
			code, err := parseInt(part)
			if err != nil {
				continue
			}
			codes = append(codes, code)
		}
	}

	return codes, nil
}

// parseInt 安全解析整数
func parseInt(s string) (int, error) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("parseInt panic recovered: %v", r)
		}
	}()

	var result int
	_, err := fmt.Sscanf(s, "%d", &result)
	return result, err
}

// defaultConfigINI 内置默认配置（完整的INI格式）
const defaultConfigINI = `[general]
threads = 25
async = false
recursive = false
deep-recursive = false
force-recursive = false
max-recursion-depth = 3
random-user-agents = false
max-time = 0
exit-on-error = false
include-status = []
exclude-status = []
exclude-sizes = []
exclude-text = []
exclude-regex = []
exclude-redirect = []
exclude-response = []
skip-on-status = []
min-response-size = 0
max-response-size = 0

[dictionary]
default-extensions = []
force-extensions = false
overwrite-extensions = false
lowercase = false
uppercase = false
capitalization = false
exclude-extensions = []
prefixes = []
suffixes = []
wordlists = []
type = file
path = ""
url = ""
db-host = ""
db-port = 3306
db-user = ""
db-password = ""
db-name = ""
db-table = wordlists
db-column = word

[request]
http-method = GET
follow-redirects = false
headers-file = ""
user-agent = ""
cookie = ""
data = ""
data-file = ""
headers = []
auth = ""
auth-type = ""

[connection]
timeout = 7.5
delay = 0
max-rate = 0
max-retries = 3
domain-check-timeout = 60
domain-check-retries = 3
headless-timeout = 30
headless-concurrency = 5
scheme = ""
proxy = ""
proxy-file = ""
replay-proxy = ""
proxies = []

[advanced]
crawl = false

[view]
full-url = false
quiet-mode = false
color = true
show-redirects-history = false
real-time-status = false
headless = false
show-all-status = false
recursive-scan = false

[output]
report-format = plain
autosave-report = false
autosave-report-folder = ""
log-file = ""
log-file-size = 0
`
