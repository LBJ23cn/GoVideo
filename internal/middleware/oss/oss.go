package oss

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	aliyunoss "github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/kiritosuki/GoVideo/internal/config"
)

const (
	defaultTimeout = 300 * time.Second
	// 超过该大小的文件走分片上传 避免单个请求过大且支持并发/断点续传
	multipartThreshold int64 = 100 << 20 // 100MB
	// 分片大小 必须是100KB~5GB之间
	defaultPartSize int64 = 10 << 20 // 10MB
)

// Client 封装阿里云OSS客户端 外部通过该对象完成对象存储操作
type Client struct {
	bucket        *aliyunoss.Bucket
	bucketBaseURL string
	publicBaseURL string
}

// NewClient 创建OSS客户端
func NewClient(cfg *config.OSSConfig) (*Client, error) {
	if cfg == nil {
		return nil, errors.New("oss config is nil")
	}
	if cfg.Bucket == "" {
		return nil, errors.New("oss bucket is required")
	}
	if strings.TrimSpace(cfg.Endpoint) == "" && strings.TrimSpace(cfg.Region) == "" {
		return nil, errors.New("oss endpoint or region is required")
	}
	if cfg.AccessKeyID == "" || cfg.AccessKeySecret == "" {
		return nil, errors.New("oss access key id and access key secret are required")
	}

	endpoint := normalizeEndpoint(cfg.Endpoint, cfg.Region)
	bucketBaseURL, err := buildBucketBaseURL(endpoint, cfg.Bucket)
	if err != nil {
		return nil, err
	}

	timeout := defaultTimeout
	if cfg.TimeoutSeconds > 0 {
		timeout = time.Duration(cfg.TimeoutSeconds) * time.Second
	}
	options := []aliyunoss.ClientOption{
		aliyunoss.Timeout(int64(timeout/time.Second), int64(timeout/time.Second)),
	}
	// 非阿里云默认域名按自定义CNAME域名处理 SDK才能正确拼接上传地址
	if !isAliyunEndpoint(endpoint) {
		options = append(options, aliyunoss.UseCname(true))
	}
	client, err := aliyunoss.New(endpoint, cfg.AccessKeyID, cfg.AccessKeySecret, options...)
	if err != nil {
		return nil, err
	}
	bucket, err := client.Bucket(cfg.Bucket)
	if err != nil {
		return nil, err
	}

	return &Client{
		bucket:        bucket,
		bucketBaseURL: bucketBaseURL,
		publicBaseURL: strings.TrimRight(strings.TrimSpace(cfg.PublicBaseURL), "/"),
	}, nil
}

// UploadFile 上传本地文件 返回对象可访问URL
// 小文件使用简单上传 大文件自动切换为分片上传
func (c *Client) UploadFile(ctx context.Context, key string, filePath string) (string, error) {
	key = normalizeKey(key)
	if c == nil || c.bucket == nil {
		return "", errors.New("oss client is not initialized")
	}
	if key == "" {
		return "", errors.New("oss object key is required")
	}
	if strings.TrimSpace(filePath) == "" {
		return "", errors.New("oss upload file path is required")
	}
	// OSS SDK 不支持 context 取消 上传前检查一次 避免已取消的请求继续上传
	if err := ctx.Err(); err != nil {
		return "", err
	}
	info, err := os.Stat(filePath)
	if err != nil {
		return "", err
	}
	if info.Size() > multipartThreshold {
		err = c.bucket.UploadFile(key, filePath, defaultPartSize)
	} else {
		err = c.bucket.PutObjectFromFile(key, filePath)
	}
	if err != nil {
		return "", err
	}
	return c.ObjectURL(key)
}

// DeleteObject 删除对象
func (c *Client) DeleteObject(ctx context.Context, key string) error {
	key = normalizeKey(key)
	if c == nil || c.bucket == nil {
		return errors.New("oss client is not initialized")
	}
	if key == "" {
		return errors.New("oss object key is required")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	return c.bucket.DeleteObject(key)
}

// ObjectURL 根据对象key生成可访问URL
func (c *Client) ObjectURL(key string) (string, error) {
	key = normalizeKey(key)
	if key == "" {
		return "", errors.New("oss object key is required")
	}
	if c == nil || c.bucketBaseURL == "" {
		return "", errors.New("oss client is not initialized")
	}
	baseURL := c.bucketBaseURL
	if c.publicBaseURL != "" {
		baseURL = c.publicBaseURL
	}
	return strings.TrimRight(baseURL, "/") + "/" + key, nil
}

// normalizeEndpoint 补全endpoint协议头 未配置时根据region拼接默认endpoint
func normalizeEndpoint(endpoint, region string) string {
	endpoint = strings.TrimRight(strings.TrimSpace(endpoint), "/")
	if endpoint == "" {
		return fmt.Sprintf("https://oss-%s.aliyuncs.com", strings.TrimSpace(region))
	}
	if !strings.HasPrefix(endpoint, "http://") && !strings.HasPrefix(endpoint, "https://") {
		endpoint = "https://" + endpoint
	}
	return endpoint
}

// isAliyunEndpoint 判断是否为阿里云默认endpoint
func isAliyunEndpoint(endpoint string) bool {
	u, err := url.Parse(endpoint)
	if err != nil {
		return false
	}
	host := u.Hostname()
	return host == "aliyuncs.com" || strings.HasSuffix(host, ".aliyuncs.com")
}

// buildBucketBaseURL 拼接bucket访问域名
// 阿里云默认endpoint需要加bucket前缀 自定义域名(endpoint为CNAME)则直接使用
func buildBucketBaseURL(endpoint, bucket string) (string, error) {
	u, err := url.Parse(endpoint)
	if err != nil || u.Host == "" {
		return "", fmt.Errorf("invalid oss endpoint %q", endpoint)
	}
	host := u.Host
	if strings.HasSuffix(host, ".aliyuncs.com") && !strings.HasPrefix(host, bucket+".") {
		host = bucket + "." + host
	}
	return u.Scheme + "://" + host, nil
}

// normalizeKey 去掉key前后的空格和前导的"/"
func normalizeKey(key string) string {
	return strings.TrimLeft(strings.TrimSpace(key), "/")
}
