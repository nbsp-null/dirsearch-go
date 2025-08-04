package cmd

import (
	"dirsearch-go/internal/config"
	"dirsearch-go/internal/scanner"
	"dirsearch-go/internal/utils"
	"dirsearch-go/internal/view"
	"fmt"

	"github.com/spf13/cobra"
)

var (
	// 全局选项
	urls        []string
	urlsFile    string
	stdin       bool
	cidr        string
	rawFile     string
	nmapReport  string
	sessionFile string
	configFile  string

	// 字典设置
	wordlists           []string
	extensions          []string
	forceExtensions     bool
	overwriteExtensions bool
	excludeExtensions   []string
	removeExtensions    bool
	prefixes            []string
	suffixes            []string
	uppercase           bool
	lowercase           bool
	capital             bool

	// Wordlist源设置
	wordlistSource     string
	wordlistURL        string
	wordlistDBHost     string
	wordlistDBPort     int
	wordlistDBUser     string
	wordlistDBPassword string
	wordlistDBName     string
	wordlistDBTable    string
	wordlistDBColumn   string

	// 通用设置
	threads           int
	async             bool
	recursive         bool
	deepRecursive     bool
	forceRecursive    bool
	maxRecursionDepth int
	recursionStatus   []string
	subdirs           []string
	excludeSubdirs    []string
	includeStatus     []string
	excludeStatus     []string
	statusFilter      string
	excludeSizes      []string
	excludeText       []string
	excludeRegex      []string
	excludeRedirect   []string
	excludeResponse   []string
	skipOnStatus      []string
	minResponseSize   int
	maxResponseSize   int
	maxTime           int
	exitOnError       bool

	// 请求设置
	httpMethod      string
	data            string
	dataFile        string
	headers         []string
	headersFile     string
	followRedirects bool
	randomAgent     bool
	auth            string
	authType        string
	certFile        string
	keyFile         string
	userAgent       string
	cookie          string

	// 连接设置
	timeout       float64
	delay         float64
	proxy         string
	proxiesFile   string
	proxyAuth     string
	replayProxy   string
	tor           bool
	scheme        string
	maxRate       int
	retries       int
	ip            string
	interfaceName string

	// 高级设置
	crawl bool

	// 视图设置
	fullURL          bool
	redirectsHistory bool
	noColor          bool
	quietMode        bool
	realTimeStatus   bool
	headless         bool
	showAllStatus    bool
	recursiveScan    bool

	// 输出设置
	output  string
	format  string
	logFile string
)

// rootCmd 根命令
var rootCmd = &cobra.Command{
	Use:   "dirsearch-go",
	Short: "An advanced web path brute-forcer",
	Long: `dirsearch-go is an advanced web path brute-forcer written in Go.

It can discover hidden files and directories on web servers by brute-forcing
common paths and extensions.`,

	RunE: func(cmd *cobra.Command, args []string) error {
		// 验证必需参数
		if len(urls) == 0 && urlsFile == "" && !stdin && cidr == "" && rawFile == "" && nmapReport == "" {
			return fmt.Errorf("URL target is missing, try using -u <url>")
		}

		if len(wordlists) == 0 {
			return fmt.Errorf("no wordlist was provided, try using -w <wordlist>")
		}

		if threads < 1 {
			return fmt.Errorf("threads number must be greater than zero")
		}

		// 启动扫描器
		return runScanner()
	},
}

// Execute 执行根命令
func Execute() error {
	return rootCmd.Execute()
}

// init 初始化命令和标志
func init() {
	// 必需参数
	rootCmd.Flags().StringArrayVarP(&urls, "url", "u", nil, "Target URL(s), can use multiple flags")
	rootCmd.Flags().StringVarP(&urlsFile, "urls-file", "l", "", "URL list file")
	rootCmd.Flags().BoolVar(&stdin, "stdin", false, "Read URL(s) from STDIN")
	rootCmd.Flags().StringVar(&cidr, "cidr", "", "Target CIDR")
	rootCmd.Flags().StringVar(&rawFile, "raw", "", "Load raw HTTP request from file")
	rootCmd.Flags().StringVar(&nmapReport, "nmap-report", "", "Load targets from nmap report")
	rootCmd.Flags().StringVarP(&sessionFile, "session", "s", "", "Session file")
	rootCmd.Flags().StringVar(&configFile, "config", "", "Path to configuration file")

	// 字典设置
	rootCmd.Flags().StringArrayVarP(&wordlists, "wordlists", "w", nil, "Wordlist files or directories contain wordlists")
	rootCmd.Flags().StringArrayVarP(&extensions, "extensions", "e", nil, "Extension list separated by commas (e.g. php,asp)")
	rootCmd.Flags().BoolVarP(&forceExtensions, "force-extensions", "f", false, "Add extensions to the end of every wordlist entry")
	rootCmd.Flags().BoolVarP(&overwriteExtensions, "overwrite-extensions", "O", false, "Overwrite other extensions in the wordlist")
	rootCmd.Flags().StringArrayVar(&excludeExtensions, "exclude-extensions", nil, "Exclude extension list separated by commas")
	rootCmd.Flags().BoolVar(&removeExtensions, "remove-extensions", false, "Remove extensions in all paths")
	rootCmd.Flags().StringArrayVar(&prefixes, "prefixes", nil, "Add custom prefixes to all wordlist entries")
	rootCmd.Flags().StringArrayVar(&suffixes, "suffixes", nil, "Add custom suffixes to all wordlist entries")
	rootCmd.Flags().BoolVarP(&uppercase, "uppercase", "U", false, "Uppercase wordlist")
	rootCmd.Flags().BoolVarP(&lowercase, "lowercase", "L", false, "Lowercase wordlist")
	rootCmd.Flags().BoolVarP(&capital, "capital", "C", false, "Capital wordlist")

	// Wordlist源设置
	rootCmd.Flags().StringVar(&wordlistSource, "wordlist-source", "file", "Wordlist source type (file, url, database)")
	rootCmd.Flags().StringVar(&wordlistURL, "wordlist-url", "", "URL to fetch wordlist from")
	rootCmd.Flags().StringVar(&wordlistDBHost, "wordlist-db-host", "", "Database host for wordlist")
	rootCmd.Flags().IntVar(&wordlistDBPort, "wordlist-db-port", 3306, "Database port for wordlist")
	rootCmd.Flags().StringVar(&wordlistDBUser, "wordlist-db-user", "", "Database user for wordlist")
	rootCmd.Flags().StringVar(&wordlistDBPassword, "wordlist-db-password", "", "Database password for wordlist")
	rootCmd.Flags().StringVar(&wordlistDBName, "wordlist-db-name", "", "Database name for wordlist")
	rootCmd.Flags().StringVar(&wordlistDBTable, "wordlist-db-table", "wordlists", "Database table for wordlist")
	rootCmd.Flags().StringVar(&wordlistDBColumn, "wordlist-db-column", "word", "Database column for wordlist")

	// 通用设置
	rootCmd.Flags().IntVarP(&threads, "threads", "t", 25, "Number of threads")
	rootCmd.Flags().BoolVar(&async, "async", false, "Enable asynchronous mode")
	rootCmd.Flags().BoolVarP(&recursive, "recursive", "r", false, "Brute-force recursively")
	rootCmd.Flags().BoolVar(&deepRecursive, "deep-recursive", false, "Perform recursive scan on every directory depth")
	rootCmd.Flags().BoolVar(&forceRecursive, "force-recursive", false, "Do recursive brute-force for every found path")
	rootCmd.Flags().IntVarP(&maxRecursionDepth, "max-recursion-depth", "R", 0, "Maximum recursion depth")
	rootCmd.Flags().StringArrayVar(&recursionStatus, "recursion-status", nil, "Valid status codes to perform recursive scan")
	rootCmd.Flags().StringArrayVar(&subdirs, "subdirs", nil, "Scan sub-directories of the given URL[s]")
	rootCmd.Flags().StringArrayVar(&excludeSubdirs, "exclude-subdirs", nil, "Exclude the following subdirectories during recursive scan")
	rootCmd.Flags().StringArrayVarP(&includeStatus, "include-status", "i", nil, "Include status codes, separated by commas")
	rootCmd.Flags().StringArrayVarP(&excludeStatus, "exclude-status", "x", nil, "Exclude status codes, separated by commas")
	rootCmd.Flags().StringVar(&statusFilter, "status", "", "Filter results by status code (e.g. 200,404)")
	rootCmd.Flags().StringArrayVar(&excludeSizes, "exclude-sizes", nil, "Exclude responses by sizes, separated by commas")
	rootCmd.Flags().StringArrayVar(&excludeText, "exclude-text", nil, "Exclude responses by text")
	rootCmd.Flags().StringArrayVar(&excludeRegex, "exclude-regex", nil, "Exclude responses by regular expression")
	rootCmd.Flags().StringArrayVar(&excludeRedirect, "exclude-redirect", nil, "Exclude responses if this regex matches redirect URL")
	rootCmd.Flags().StringArrayVar(&excludeResponse, "exclude-response", nil, "Exclude responses similar to response of this page")
	rootCmd.Flags().StringArrayVar(&skipOnStatus, "skip-on-status", nil, "Skip target whenever hit one of these status codes")
	rootCmd.Flags().IntVar(&minResponseSize, "min-response-size", 0, "Minimum response length")
	rootCmd.Flags().IntVar(&maxResponseSize, "max-response-size", 0, "Maximum response length")
	rootCmd.Flags().IntVar(&maxTime, "max-time", 0, "Maximum runtime for the scan")
	rootCmd.Flags().BoolVar(&exitOnError, "exit-on-error", false, "Exit whenever an error occurs")

	// 请求设置
	rootCmd.Flags().StringVarP(&httpMethod, "http-method", "m", "GET", "HTTP method (default: GET)")
	rootCmd.Flags().StringVarP(&data, "data", "d", "", "HTTP request data")
	rootCmd.Flags().StringVar(&dataFile, "data-file", "", "File contains HTTP request data")
	rootCmd.Flags().StringArrayVarP(&headers, "header", "H", nil, "HTTP request header, can use multiple flags")
	rootCmd.Flags().StringVar(&headersFile, "headers-file", "", "File contains HTTP request headers")
	rootCmd.Flags().BoolVarP(&followRedirects, "follow-redirects", "F", false, "Follow HTTP redirects")
	rootCmd.Flags().BoolVar(&randomAgent, "random-agent", false, "Choose a random User-Agent for each request")
	rootCmd.Flags().StringVar(&auth, "auth", "", "Authentication credential (e.g. user:password or bearer token)")
	rootCmd.Flags().StringVar(&authType, "auth-type", "", "Authentication type (basic, digest, bearer, ntlm, jwt)")
	rootCmd.Flags().StringVar(&certFile, "cert-file", "", "File contains client-side certificate")
	rootCmd.Flags().StringVar(&keyFile, "key-file", "", "File contains client-side certificate private key")
	rootCmd.Flags().StringVar(&userAgent, "user-agent", "", "User-Agent")
	rootCmd.Flags().StringVar(&cookie, "cookie", "", "Cookie")

	// 连接设置
	rootCmd.Flags().Float64Var(&timeout, "timeout", 7.5, "Connection timeout")
	rootCmd.Flags().Float64Var(&delay, "delay", 0, "Delay between requests")
	rootCmd.Flags().StringVarP(&proxy, "proxy", "p", "", "Proxy URL (HTTP/SOCKS), can use multiple flags")
	rootCmd.Flags().StringVar(&proxiesFile, "proxies-file", "", "File contains proxy servers")
	rootCmd.Flags().StringVar(&proxyAuth, "proxy-auth", "", "Proxy authentication credential")
	rootCmd.Flags().StringVar(&replayProxy, "replay-proxy", "", "Proxy to replay with found paths")
	rootCmd.Flags().BoolVar(&tor, "tor", false, "Use Tor network as proxy")
	rootCmd.Flags().StringVar(&scheme, "scheme", "", "Scheme for raw request or if there is no scheme in the URL")
	rootCmd.Flags().IntVar(&maxRate, "max-rate", 0, "Max requests per second")
	rootCmd.Flags().IntVar(&retries, "retries", 1, "Number of retries for failed requests")
	rootCmd.Flags().StringVar(&ip, "ip", "", "Server IP address")
	rootCmd.Flags().StringVar(&interfaceName, "interface", "", "Network interface to use")

	// 高级设置
	rootCmd.Flags().BoolVar(&crawl, "crawl", false, "Crawl for new paths in responses")

	// 视图设置
	rootCmd.Flags().BoolVar(&fullURL, "full-url", false, "Full URLs in the output")
	rootCmd.Flags().BoolVar(&redirectsHistory, "redirects-history", false, "Show redirects history")
	rootCmd.Flags().BoolVar(&noColor, "no-color", false, "No colored output")
	rootCmd.Flags().BoolVarP(&quietMode, "quiet-mode", "q", false, "Quiet mode")
	rootCmd.Flags().BoolVar(&realTimeStatus, "real-time-status", false, "Enable real-time status display")
	rootCmd.Flags().BoolVar(&headless, "headless", false, "Use headless browser for scanning")
	rootCmd.Flags().BoolVar(&showAllStatus, "show-all-status", false, "Show all status codes (default: only 200 and 403)")
	rootCmd.Flags().BoolVar(&recursiveScan, "recursive-scan", false, "Enable recursive scanning for directories (200/403)")

	// 输出设置
	rootCmd.Flags().StringVarP(&output, "output", "o", "", "Output file or MySQL/PostgreSQL URL")
	rootCmd.Flags().StringVar(&format, "format", "plain", "Report format (Available: simple, plain, json, xml, md, csv, html, sqlite, mysql, postgresql)")
	rootCmd.Flags().StringVar(&logFile, "log", "", "Log file")

	// 版本信息
	rootCmd.Flags().Bool("version", false, "Show program's version number and exit")
}

// runScanner 运行扫描器
func runScanner() error {
	// 获取配置
	cfg := config.GetConfig()
	if cfg == nil {
		return fmt.Errorf("failed to get configuration")
	}

	// 处理目标URL
	var targets []string

	// 从命令行参数获取URL
	if len(urls) > 0 {
		targets = append(targets, urls...)
	}

	// 从文件读取URL
	if urlsFile != "" {
		fileTargets, err := utils.ReadLinesFromFile(urlsFile)
		if err != nil {
			return fmt.Errorf("failed to read URLs file: %w", err)
		}
		targets = append(targets, fileTargets...)
	}

	// 从标准输入读取URL
	if stdin {
		stdinTargets, err := utils.ReadLinesFromStdin()
		if err != nil {
			return fmt.Errorf("failed to read URLs from stdin: %w", err)
		}
		targets = append(targets, stdinTargets...)
	}

	// 从CIDR解析URL
	if cidr != "" {
		cidrTargets, err := utils.ParseCIDR(cidr)
		if err != nil {
			return fmt.Errorf("failed to parse CIDR: %w", err)
		}
		targets = append(targets, cidrTargets...)
	}

	if len(targets) == 0 {
		return fmt.Errorf("no targets specified")
	}

	// 清理和验证URL
	var cleanTargets []string
	for _, target := range targets {
		cleanTarget := utils.CleanURL(target)
		if utils.IsValidURL(cleanTarget) {
			cleanTargets = append(cleanTargets, cleanTarget)
		}
	}

	if len(cleanTargets) == 0 {
		return fmt.Errorf("no valid targets found")
	}

	// 更新配置
	updateConfigFromFlags(cfg)

	// 创建扫描器
	scanner, err := scanner.NewScanner(cfg)
	if err != nil {
		return fmt.Errorf("failed to create scanner: %w", err)
	}

	// 开始扫描
	fmt.Printf("Starting scan with %d targets and %d threads...\n", len(cleanTargets), cfg.General.Threads)

	// 执行扫描并获取结果
	results, err := scanner.Scan(cleanTargets)
	if err != nil {
		return fmt.Errorf("scan failed: %w", err)
	}

	fmt.Printf("Scan completed. Found %d results.\n", len(results))

	// 保存结果
	if output != "" {
		if err := scanner.SaveResults(output); err != nil {
			return fmt.Errorf("failed to save results: %w", err)
		}
		fmt.Printf("Results saved to: %s\n", output)
	}

	// 显示结果
	displayResults(results)

	return nil
}

// updateConfigFromFlags 从命令行标志更新配置
func updateConfigFromFlags(cfg *config.Config) {
	// 更新字典配置
	if len(wordlists) > 0 {
		cfg.Dictionary.Wordlists = wordlists
	}
	if len(extensions) > 0 {
		cfg.Dictionary.DefaultExtensions = extensions
	}
	if forceExtensions {
		cfg.Dictionary.ForceExtensions = true
	}
	if overwriteExtensions {
		cfg.Dictionary.OverwriteExtensions = true
	}
	if len(excludeExtensions) > 0 {
		cfg.Dictionary.ExcludeExtensions = excludeExtensions
	}
	if removeExtensions {
		// TODO: 实现移除扩展名功能
	}
	if len(prefixes) > 0 {
		cfg.Dictionary.Prefixes = prefixes
	}
	if len(suffixes) > 0 {
		cfg.Dictionary.Suffixes = suffixes
	}
	if uppercase {
		cfg.Dictionary.Uppercase = true
	}
	if lowercase {
		cfg.Dictionary.Lowercase = true
	}
	if capital {
		cfg.Dictionary.Capitalization = true
	}

	// 更新wordlist源配置
	if wordlistSource != "" {
		cfg.Dictionary.Source.Type = wordlistSource
	}
	if wordlistURL != "" {
		cfg.Dictionary.Source.URL = wordlistURL
	}
	if wordlistDBHost != "" {
		cfg.Dictionary.Source.DBHost = wordlistDBHost
	}
	if wordlistDBPort > 0 {
		cfg.Dictionary.Source.DBPort = wordlistDBPort
	}
	if wordlistDBUser != "" {
		cfg.Dictionary.Source.DBUser = wordlistDBUser
	}
	if wordlistDBPassword != "" {
		cfg.Dictionary.Source.DBPass = wordlistDBPassword
	}
	if wordlistDBName != "" {
		cfg.Dictionary.Source.DBName = wordlistDBName
	}
	if wordlistDBTable != "" {
		cfg.Dictionary.Source.DBTable = wordlistDBTable
	}
	if wordlistDBColumn != "" {
		cfg.Dictionary.Source.DBColumn = wordlistDBColumn
	}

	// 更新通用配置
	if threads > 0 {
		cfg.General.Threads = threads
	}
	if async {
		cfg.General.Async = true
	}
	if recursive {
		cfg.General.Recursive = true
	}
	if deepRecursive {
		cfg.General.DeepRecursive = true
	}
	if forceRecursive {
		cfg.General.ForceRecursive = true
	}
	if maxRecursionDepth > 0 {
		cfg.General.MaxRecursionDepth = maxRecursionDepth
	}
	if len(recursionStatus) > 0 {
		cfg.General.RecursionStatus = recursionStatus
	}
	if len(subdirs) > 0 {
		// TODO: 实现子目录扫描
	}
	if len(excludeSubdirs) > 0 {
		cfg.General.ExcludeSubdirs = excludeSubdirs
	}
	if len(includeStatus) > 0 {
		cfg.General.IncludeStatus = includeStatus
	}
	if len(excludeStatus) > 0 {
		cfg.General.ExcludeStatus = excludeStatus
	}
	if len(excludeSizes) > 0 {
		cfg.General.ExcludeSizes = excludeSizes
	}
	if len(excludeText) > 0 {
		cfg.General.ExcludeText = excludeText
	}
	if len(excludeRegex) > 0 {
		cfg.General.ExcludeRegex = excludeRegex
	}
	if len(excludeRedirect) > 0 {
		cfg.General.ExcludeRedirect = excludeRedirect
	}
	if len(excludeResponse) > 0 {
		cfg.General.ExcludeResponse = excludeResponse
	}
	if len(skipOnStatus) > 0 {
		cfg.General.SkipOnStatus = skipOnStatus
	}
	if minResponseSize > 0 {
		cfg.General.MinResponseSize = minResponseSize
	}
	if maxResponseSize > 0 {
		cfg.General.MaxResponseSize = maxResponseSize
	}
	if maxTime > 0 {
		cfg.General.MaxTime = maxTime
	}
	if exitOnError {
		cfg.General.ExitOnError = true
	}

	// 更新请求配置
	if httpMethod != "" {
		cfg.Request.HTTPMethod = httpMethod
	}
	if data != "" {
		cfg.Request.Data = data
	}
	if dataFile != "" {
		cfg.Request.DataFile = dataFile
	}
	if len(headers) > 0 {
		cfg.Request.Headers = headers
	}
	if headersFile != "" {
		cfg.Request.HeadersFile = headersFile
	}
	if followRedirects {
		cfg.Request.FollowRedirects = true
	}
	if randomAgent {
		cfg.General.RandomUserAgents = true
	}
	if userAgent != "" {
		cfg.Request.UserAgent = userAgent
	}
	if cookie != "" {
		cfg.Request.Cookie = cookie
	}

	// 更新连接配置
	if timeout > 0 {
		cfg.Connection.Timeout = timeout
	}
	if delay > 0 {
		cfg.Connection.Delay = delay
	}
	if proxy != "" {
		cfg.Connection.Proxy = proxy
	}
	if proxiesFile != "" {
		cfg.Connection.ProxyFile = proxiesFile
	}
	if proxyAuth != "" {
		// TODO: 实现代理认证
	}
	if replayProxy != "" {
		cfg.Connection.ReplayProxy = replayProxy
	}
	if tor {
		// TODO: 实现Tor代理
	}
	if scheme != "" {
		cfg.Connection.Scheme = scheme
	}
	if maxRate > 0 {
		cfg.Connection.MaxRate = maxRate
	}
	if retries > 0 {
		cfg.Connection.MaxRetries = retries
	}
	if ip != "" {
		// TODO: 实现IP绑定
	}
	if interfaceName != "" {
		// TODO: 实现网络接口绑定
	}

	// 更新高级配置
	if crawl {
		cfg.Advanced.Crawl = true
	}

	// 更新视图配置
	if fullURL {
		cfg.View.FullURL = true
	}
	if redirectsHistory {
		cfg.View.ShowRedirectsHistory = true
	}
	if noColor {
		cfg.View.Color = false
	}
	if quietMode {
		cfg.View.QuietMode = true
	}
	if realTimeStatus {
		cfg.View.RealTimeStatus = true
	}
	if headless {
		cfg.View.Headless = true
	}
	if showAllStatus {
		cfg.View.ShowAllStatus = true
	}
	if recursiveScan {
		cfg.View.RecursiveScan = true
	}

	// 更新输出配置
	if output != "" {
		// TODO: 实现输出配置
	}
	if format != "" {
		cfg.Output.ReportFormat = format
	}
	if logFile != "" {
		cfg.Output.LogFile = logFile
	}
}

// displayResults 显示扫描结果
func displayResults(results []scanner.ScanResult) {
	if len(results) == 0 {
		fmt.Println("No results found.")
		return
	}

	// 如果是无头模式，不显示详细结果
	if headless {
		return
	}

	// 创建颜色管理器
	colorManager := view.NewColorManager(!noColor)

	// 状态码筛选
	if statusFilter != "" {
		results = filterResultsByStatus(results, statusFilter)
	}

	// 过滤结果：默认只显示200和403，除非指定显示所有状态码
	var filteredResults []scanner.ScanResult
	cfg := config.GetConfig()

	if cfg != nil && cfg.View.ShowAllStatus {
		filteredResults = results
	} else {
		// 默认只显示200和403
		for _, result := range results {
			if result.StatusCode == 200 || result.StatusCode == 403 {
				filteredResults = append(filteredResults, result)
			}
		}
	}

	if len(filteredResults) == 0 {
		fmt.Println("No results found with status codes 200 or 403.")
		return
	}

	// 显示状态码统计
	displayStatusSummary(filteredResults, colorManager)

	fmt.Println("\nScan Results:")
	fmt.Println("=============")

	for _, result := range filteredResults {
		coloredStatus := colorManager.ColorizeStatus(result.StatusCode)
		coloredURL := colorManager.ColorizeURL(result.URL)
		coloredPath := colorManager.ColorizeURL(result.Path)

		fmt.Printf("[%s] %s%s\n", coloredStatus, coloredURL, coloredPath)

		if result.Title != "" {
			coloredTitle := colorManager.ColorizeTitle(result.Title)
			fmt.Printf("    Title: %s\n", coloredTitle)
		}
		if result.Redirect != "" {
			coloredRedirect := colorManager.ColorizeRedirect(result.Redirect)
			fmt.Printf("    Redirect: %s\n", coloredRedirect)
		}
		if result.Error != nil {
			coloredError := colorManager.ColorizeError(result.Error.Error())
			fmt.Printf("    Error: %s\n", coloredError)
		}
		fmt.Println()
	}
}

// filterResultsByStatus 根据状态码筛选结果
func filterResultsByStatus(results []scanner.ScanResult, statusFilter string) []scanner.ScanResult {
	var filtered []scanner.ScanResult
	statusCodes, err := config.ParseStatusCodes(statusFilter)
	if err != nil {
		fmt.Printf("Warning: Invalid status filter '%s': %v\n", statusFilter, err)
		return results
	}

	for _, result := range results {
		for _, code := range statusCodes {
			if result.StatusCode == code {
				filtered = append(filtered, result)
				break
			}
		}
	}

	return filtered
}

// displayStatusSummary 显示状态码统计
func displayStatusSummary(results []scanner.ScanResult, colorManager *view.ColorManager) {
	statusCount := make(map[int]int)
	for _, result := range results {
		statusCount[result.StatusCode]++
	}

	fmt.Println("\nStatus Code Summary:")
	fmt.Println("====================")
	for code, count := range statusCount {
		coloredCode := colorManager.ColorizeStatus(code)
		fmt.Printf("[%s]: %d results\n", coloredCode, count)
	}
	fmt.Println()
}
