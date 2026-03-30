package core

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Config struct {
	Endpoint  string
	Region    string
	Bucket    string
	AccessKey string
	SecretKey string
	Prefix    string
}

type S3Manager struct {
	client *s3.Client
	bucket string
	prefix string
}

type S3Entry struct {
	Name  string `json:"name"`
	Path  string `json:"path"`
	IsDir bool   `json:"isDir"`
	Size  int64  `json:"size"`
}

var GlobalS3Manager *S3Manager

func InitS3(cfg S3Config) error {
	if cfg.Region == "" {
		cfg.Region = "us-east-1"
	}

	client := s3.New(s3.Options{
		Region: cfg.Region,
		Credentials: credentials.NewStaticCredentialsProvider(
			cfg.AccessKey,
			cfg.SecretKey,
			"",
		),
		BaseEndpoint: aws.String(cfg.Endpoint),
		UsePathStyle: true,
	})

	GlobalS3Manager = &S3Manager{
		client: client,
		bucket: cfg.Bucket,
		prefix: cfg.Prefix,
	}
	return nil
}

func (m *S3Manager) ListObjects(prefix string) ([]S3Entry, error) {
	fullPrefix := m.prefix + prefix

	input := &s3.ListObjectsV2Input{
		Bucket:    aws.String(m.bucket),
		Prefix:    aws.String(fullPrefix),
		Delimiter: aws.String("/"),
	}

	result, err := m.client.ListObjectsV2(context.Background(), input)
	if err != nil {
		return nil, fmt.Errorf("failed to list objects: %v", err)
	}

	var entries []S3Entry

	// Add directories (common prefixes)
	for _, cp := range result.CommonPrefixes {
		dirPath := strings.TrimPrefix(*cp.Prefix, m.prefix)
		name := strings.TrimSuffix(strings.TrimPrefix(*cp.Prefix, fullPrefix), "/")
		if name == "" {
			continue
		}
		entries = append(entries, S3Entry{
			Name:  name,
			Path:  dirPath,
			IsDir: true,
		})
	}

	// Add files (objects)
	for _, obj := range result.Contents {
		filePath := strings.TrimPrefix(*obj.Key, m.prefix)
		name := strings.TrimPrefix(*obj.Key, fullPrefix)
		if name == "" || strings.HasSuffix(name, "/") {
			continue
		}
		entries = append(entries, S3Entry{
			Name:  name,
			Path:  filePath,
			IsDir: false,
			Size:  *obj.Size,
		})
	}

	return entries, nil
}

func (m *S3Manager) ReadObject(key string) (string, error) {
	// Validate the key is within the configured prefix
	fullKey := m.prefix + key
	if !strings.HasPrefix(fullKey, m.prefix) {
		return "", fmt.Errorf("access denied: key is outside allowed prefix")
	}

	input := &s3.GetObjectInput{
		Bucket: aws.String(m.bucket),
		Key:    aws.String(fullKey),
	}

	result, err := m.client.GetObject(context.Background(), input)
	if err != nil {
		return "", fmt.Errorf("failed to read object: %v", err)
	}
	defer result.Body.Close()

	data, err := io.ReadAll(result.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read object body: %v", err)
	}

	return string(data), nil
}
