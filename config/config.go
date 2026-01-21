package config

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/caarlos0/env/v6"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/yaoapp/kun/exception"
	"github.com/yaoapp/kun/log"
	"github.com/yaoapp/yao/crypto"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Conf 配置参数
var Conf Config

// LogOutput 日志输出
var LogOutput io.WriteCloser // 日志文件

// DSLExtensions the dsl file Extensions
var DSLExtensions = []string{"*.yao", "*.json", "*.jsonc"}

func init() {
	Init()
}

// Init setting
func Init() {
	// Determine app root: YAO_ROOT env > find app.yao > current directory
	root := os.Getenv("YAO_ROOT")
	if root == "" {
		root = findAppRoot()
	}
	if root == "" {
		root = "."
	}

	filename, _ := filepath.Abs(filepath.Join(root, ".env"))
	if _, err := os.Stat(filename); errors.Is(err, os.ErrNotExist) {
		Conf = LoadWithRoot(root)
		ApplyMode()
		return
	}

	// Load .env then override root if auto-detected
	Conf = LoadFromWithRoot(filename, root)
	ApplyMode()
}

// ApplyMode applies production or development mode based on Conf.Mode
func ApplyMode() {
	switch Conf.Mode {
	case "production":
		Production()
	case "development":
		Development()
	}
}

// findAppRoot finds the Yao application root directory by looking for app.yao
// It traverses up from the current directory until it finds app.yao or reaches the filesystem root
func findAppRoot() string {
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}

	for {
		// Check for app.yao, app.json, or app.jsonc
		for _, appFile := range []string{"app.yao", "app.json", "app.jsonc"} {
			appFilePath := filepath.Join(dir, appFile)
			if _, err := os.Stat(appFilePath); err == nil {
				return dir
			}
		}

		// Move to parent directory
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached root, no app.yao found
			break
		}
		dir = parent
	}

	return ""
}

// LoadFrom 从配置项中加载
func LoadFrom(envfile string) Config {
	return LoadFromWithRoot(envfile, "")
}

// LoadFromWithRoot loads config from env file with optional root override
func LoadFromWithRoot(envfile string, root string) Config {
	file, err := filepath.Abs(envfile)
	if err != nil {
		cfg := LoadWithRoot(root)
		ReloadLog()
		return cfg
	}

	// load from env
	godotenv.Overload(file)
	cfg := LoadWithRoot(root)
	ReloadLog()
	return cfg
}

// Load the config
func Load() Config {
	return LoadWithRoot("")
}

// LoadWithRoot loads config with an optional root override
// If root is empty, uses YAO_ROOT env or current directory
func LoadWithRoot(root string) Config {
	cfg := Config{}
	if err := env.Parse(&cfg); err != nil {
		exception.New("Can't read config %s", 500, err.Error()).Throw()
	}

	// Root path: use provided root > env YAO_ROOT > default "."
	if root != "" {
		cfg.Root, _ = filepath.Abs(root)
	} else {
		cfg.Root, _ = filepath.Abs(cfg.Root)
	}

	// App Root
	if cfg.AppSource == "" {
		cfg.AppSource = cfg.Root
	}

	// Studio Secret
	if cfg.Studio.Secret == "" {
		v, err := crypto.Hash(crypto.HashTypes["SHA256"], uuid.New().String())
		if err != nil {
			exception.New("Can't gengrate studio secret %s", 500, err.Error()).Throw()
		}
		cfg.Studio.Secret = strings.ToUpper(v)
		cfg.Studio.Auto = true
	}

	// DataRoot
	if cfg.DataRoot == "" {
		cfg.DataRoot = filepath.Join(cfg.Root, "data")
		if !filepath.IsAbs(cfg.DataRoot) {
			cfg.DataRoot, _ = filepath.Abs(cfg.DataRoot)
		}
	}

	// DB Primary - Resolve relative paths for SQLite only (case-insensitive)
	if strings.HasPrefix(strings.ToLower(cfg.DB.Driver), "sqlite") {
		for i, dsn := range cfg.DB.Primary {
			if dsn != "" && !filepath.IsAbs(dsn) && isValidFilePath(dsn) {
				cfg.DB.Primary[i] = filepath.Join(cfg.Root, dsn)
			}
		}
		for i, dsn := range cfg.DB.Secondary {
			if dsn != "" && !filepath.IsAbs(dsn) && isValidFilePath(dsn) {
				cfg.DB.Secondary[i] = filepath.Join(cfg.Root, dsn)
			}
		}
	}

	// Trace Driver - default based on mode
	if cfg.Trace.Driver == "" {
		if cfg.Mode == "development" {
			cfg.Trace.Driver = "local"
		} else {
			cfg.Trace.Driver = "store"
		}
	}

	// Trace Path - default to same directory as log file when using local driver
	if cfg.Trace.Driver == "local" {
		if cfg.Trace.Path == "" {
			// Use the log file directory
			logDir := cfg.GetLogDir()
			cfg.Trace.Path = filepath.Join(logDir, "traces")
		}

		if !filepath.IsAbs(cfg.Trace.Path) {
			cfg.Trace.Path = filepath.Join(cfg.Root, cfg.Trace.Path)
		}
	}

	// Trace Prefix - default prefix for store driver
	if cfg.Trace.Driver == "store" && cfg.Trace.Prefix == "" {
		cfg.Trace.Prefix = "trace:"
	}

	return cfg
}
func setLogLevel() {
	level := int(log.InfoLevel)
	switch strings.ToLower(Conf.LogLevel) {
	case "trace":
		level = int(log.TraceLevel)
	case "debug":
		level = int(log.DebugLevel)
	case "info":
		level = int(log.InfoLevel)
	case "warn":
		level = int(log.WarnLevel)
	case "error":
		level = int(log.ErrorLevel)
	case "panic":
		level = int(log.PanicLevel)
	case "fatal":
		level = int(log.FatalLevel)
	default:
		switch Conf.Mode {
		case "prodcution":
			level = int(log.ErrorLevel)
		case "development":
			level = int(log.TraceLevel)
		}
	}
	log.SetLevel(log.Level(level))
}

// GetLogDir returns the directory of the log file
func (cfg *Config) GetLogDir() string {
	logPath := cfg.Log
	if logPath == "" {
		logPath = filepath.Join(cfg.Root, "logs", "application.log")
	}

	if !filepath.IsAbs(logPath) {
		logPath = filepath.Join(cfg.Root, logPath)
	}

	return filepath.Dir(logPath)
}

// Production 设定为生产环境
func Production() {
	os.Setenv("YAO_MODE", "production")
	Conf.Mode = "production"
	// log.SetLevel(log.InfoLevel)
	setLogLevel()
	log.SetFormatter(log.TEXT)
	if Conf.LogMode == "JSON" {
		log.SetFormatter(log.JSON)
	}
	gin.SetMode(gin.ReleaseMode)
	ReloadLog()
}

// Development 设定为开发环境
func Development() {
	os.Setenv("YAO_MODE", "development")
	Conf.Mode = "development"
	// log.SetLevel(log.TraceLevel)
	setLogLevel()
	log.SetFormatter(log.TEXT)
	if Conf.LogMode == "JSON" {
		log.SetFormatter(log.JSON)
	}
	gin.SetMode(gin.DebugMode)
	ReloadLog()
}

// ReloadLog 重新打开日志
func ReloadLog() {
	CloseLog()
	OpenLog()
}

// OpenLog 打开日志
func OpenLog() {

	if Conf.Log == "" {
		Conf.Log = filepath.Join(Conf.Root, "logs", "application.log")
	}

	if !filepath.IsAbs(Conf.Log) {
		Conf.Log = filepath.Join(Conf.Root, Conf.Log)
	}

	logfile, err := filepath.Abs(Conf.Log)
	if err != nil {
		return
	}

	logpath := filepath.Dir(logfile)

	// Check if the log path exists
	if _, err := os.Stat(logpath); errors.Is(err, os.ErrNotExist) {
		LogOutput, _ := os.OpenFile(os.DevNull, os.O_WRONLY, 0666)
		log.SetOutput(LogOutput)
		gin.DefaultWriter = io.MultiWriter(LogOutput)
		return
	}

	LogOutput = &lumberjack.Logger{
		Filename:   logfile,
		MaxSize:    Conf.LogMaxSize, // megabytes
		MaxBackups: Conf.LogMaxBackups,
		MaxAge:     Conf.LogMaxAage, //days
		LocalTime:  Conf.LogLocalTime,
	}

	log.SetOutput(LogOutput)
	gin.DefaultWriter = io.MultiWriter(LogOutput)
}

// CloseLog 关闭日志
func CloseLog() {
	if LogOutput != nil {
		err := LogOutput.Close()
		if err != nil {
			log.Error("Failed to close log output: %v", err)
			return
		}
	}
}

// IsDevelopment returns true if the current mode is development
func IsDevelopment() bool {
	return Conf.Mode == "development"
}

// isValidFilePath checks if the string is a valid file path (not a DSN connection string)
func isValidFilePath(s string) bool {
	if s == "" {
		return false
	}

	if strings.Contains(s, "@") {
		return false
	}

	if strings.Contains(s, "://") {
		return false
	}

	if filepath.IsAbs(s) {
		return true
	}

	return true
}
