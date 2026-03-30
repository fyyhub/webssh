package main

import (
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
	"webssh/controller"
	"webssh/core"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
)

//go:embed public/*
var f embed.FS

var (
	port       = flag.Int("p", 8888, "服务运行端口")
	v          = flag.Bool("v", false, "显示版本号")
	authInfo   = flag.String("a", "", "开启账号密码登录验证, '-a user:pass'的格式传参")
	timeout    int
	savePass   bool
	version    string
	buildDate  string
	goVersion  string
	gitVersion string
	username   string
	password   string

	// S3 configuration
	s3Endpoint  string
	s3Region    string
	s3Bucket    string
	s3AccessKey string
	s3SecretKey string
	s3Prefix    string
	s3Enabled   bool
)

func init() {
	flag.IntVar(&timeout, "t", 120, "ssh连接超时时间(min)")
	flag.BoolVar(&savePass, "s", true, "保存ssh密码")
	if envVal, ok := os.LookupEnv("savePass"); ok {
		if b, err := strconv.ParseBool(envVal); err == nil {
			savePass = b
		}
	}
	if envVal, ok := os.LookupEnv("authInfo"); ok {
		*authInfo = envVal
	}
	if envVal, ok := os.LookupEnv("port"); ok {
		if b, err := strconv.Atoi(envVal); err == nil {
			*port = b
		}
	}
	flag.Parse()
	if *v {
		fmt.Printf("Version: %s\n\n", version)
		fmt.Printf("BuildDate: %s\n\n", buildDate)
		fmt.Printf("GoVersion: %s\n\n", goVersion)
		fmt.Printf("GitVersion: %s\n\n", gitVersion)
		os.Exit(0)
	}
	if *authInfo != "" {
		accountInfo := strings.Split(*authInfo, ":")
		if len(accountInfo) != 2 || accountInfo[0] == "" || accountInfo[1] == "" {
			fmt.Println("请按'user:pass'的格式来传参或设置环境变量, 且账号密码都不能为空!")
			os.Exit(0)
		}
		username, password = accountInfo[0], accountInfo[1]
	}

	// S3 environment variables
	s3Endpoint = os.Getenv("S3_ENDPOINT")
	s3Region = os.Getenv("S3_REGION")
	s3Bucket = os.Getenv("S3_BUCKET")
	s3AccessKey = os.Getenv("S3_ACCESS_KEY")
	s3SecretKey = os.Getenv("S3_SECRET_KEY")
	s3Prefix = os.Getenv("S3_PREFIX")
	if s3Endpoint != "" && s3Bucket != "" && s3AccessKey != "" && s3SecretKey != "" {
		s3Enabled = true
	}
}

func main() {
	// Initialize S3 if configured
	if s3Enabled {
		err := core.InitS3(core.S3Config{
			Endpoint:  s3Endpoint,
			Region:    s3Region,
			Bucket:    s3Bucket,
			AccessKey: s3AccessKey,
			SecretKey: s3SecretKey,
			Prefix:    s3Prefix,
		})
		if err != nil {
			fmt.Printf("Failed to initialize S3: %v\n", err)
		} else {
			fmt.Println("S3 key browser enabled")
		}
	}

    server := gin.New()
    server.Use(gin.Recovery())
    server.SetTrustedProxies(nil)
    server.Use(gzip.Gzip(gzip.DefaultCompression))

	// --- API Routes ---
	// No BasicAuth for API routes as per original logic.
	// If auth is needed for APIs, these routes should be moved inside the auth-enabled group below.
	server.GET("/term", func(c *gin.Context) {
		controller.TermWs(c, time.Duration(timeout)*time.Minute)
	})
	server.GET("/check", func(c *gin.Context) {
		responseBody := controller.CheckSSH(c)
		responseBody.Data = map[string]interface{}{
			"savePass": savePass,
		}
		c.JSON(200, responseBody)
	})
	file := server.Group("/file")
	{
		file.GET("/list", func(c *gin.Context) {
			c.JSON(200, controller.FileList(c))
		})
		file.GET("/download", func(c *gin.Context) {
			controller.DownloadFile(c)
		})
		file.POST("/upload", func(c *gin.Context) {
			c.JSON(200, controller.UploadFile(c))
		})
		file.GET("/progress", func(c *gin.Context) {
			controller.UploadProgressWs(c)
		})
	}

	// --- S3 Routes ---
	s3Group := server.Group("/s3")
	{
		s3Group.GET("/config", func(c *gin.Context) {
			c.JSON(200, gin.H{"enabled": s3Enabled})
		})
		s3Group.GET("/list", func(c *gin.Context) {
			c.JSON(200, controller.S3List(c))
		})
	}

	// --- Static Files & SPA Frontend ---
	// Serve static files from the 'static' directory
	staticFS, _ := fs.Sub(f, "public/static")
	server.StaticFS("/static", http.FS(staticFS))
	
	// For any other route, serve the index.html file.
	// This makes it compatible with Vue Router's history mode.
	server.NoRoute(func(c *gin.Context) {
		if *authInfo != "" {
			// If auth is enabled, check credentials.
			// This is a simplified check. For production, use a proper session/token mechanism.
			user, pass, hasAuth := c.Request.BasicAuth()
			if !hasAuth || user != username || pass != password {
				c.Header("WWW-Authenticate", `Basic realm="Restricted"`)
				c.AbortWithStatus(http.StatusUnauthorized)
				return
			}
		}
		
		indexHTML, err := f.ReadFile("public/index.html")
		if err != nil {
			c.String(http.StatusInternalServerError, "index.html not found")
			return
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", indexHTML)
	})

	fmt.Printf("Github：https://github.com/eooce/webssh\n")
	server.Run(fmt.Sprintf(":%d", *port))
}
