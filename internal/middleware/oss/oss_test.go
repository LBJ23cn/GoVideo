package oss

import (
	"context"
	"testing"

	"github.com/kiritosuki/GoVideo/internal/config"
)

// 测试OSS配置为空时创建客户端会失败
// 这个测试用于覆盖启动时配置缺失的错误路径
func TestNewClientRejectsNilConfig(t *testing.T) {
	if _, err := NewClient(nil); err == nil {
		t.Fatal("expected nil oss config to return error")
	}
}

// 测试OSS核心字段缺失时创建客户端会失败
// bucket/endpoint或region/access key 是上传对象所需的最小配置 缺失时不应该创建可用客户端
func TestNewClientRejectsMissingRequiredFields(t *testing.T) {
	tests := []struct {
		name string
		cfg  config.OSSConfig
	}{
		{name: "missing bucket", cfg: config.OSSConfig{Region: "cn-beijing", AccessKeyID: "ak", AccessKeySecret: "sk"}},
		{name: "missing endpoint and region", cfg: config.OSSConfig{Bucket: "bucket-123", AccessKeyID: "ak", AccessKeySecret: "sk"}},
		{name: "missing access key", cfg: config.OSSConfig{Bucket: "bucket-123", Region: "cn-beijing"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := NewClient(&tt.cfg); err == nil {
				t.Fatal("expected invalid oss config to return error")
			}
		})
	}
}

// 测试不配置public_base_url时 使用bucket和region自动拼接访问URL
// 这个测试不触发真实上传 只验证URL生成规则
func TestObjectURLUsesBucketURLByDefault(t *testing.T) {
	client := newTestClient(t, "")

	url, err := client.ObjectURL(" /videos/1/a.mp4 ")
	if err != nil {
		t.Fatalf("object url: %v", err)
	}
	want := "https://examplebucket.oss-cn-beijing.aliyuncs.com/videos/1/a.mp4"
	if url != want {
		t.Fatalf("expected %q, got %q", want, url)
	}
}

// 测试显式配置endpoint时优先使用endpoint拼接访问URL
func TestObjectURLUsesConfiguredEndpoint(t *testing.T) {
	client, err := NewClient(&config.OSSConfig{
		Bucket:          "examplebucket",
		Endpoint:        "oss-cn-shanghai.aliyuncs.com",
		AccessKeyID:     "ak",
		AccessKeySecret: "sk",
	})
	if err != nil {
		t.Fatalf("new oss client: %v", err)
	}

	url, err := client.ObjectURL("/covers/1/a.png")
	if err != nil {
		t.Fatalf("object url: %v", err)
	}
	want := "https://examplebucket.oss-cn-shanghai.aliyuncs.com/covers/1/a.png"
	if url != want {
		t.Fatalf("expected %q, got %q", want, url)
	}
}

// 测试自定义CNAME域名作为endpoint时不添加bucket前缀
func TestObjectURLUsesCnameEndpointWithoutBucketPrefix(t *testing.T) {
	client, err := NewClient(&config.OSSConfig{
		Bucket:          "examplebucket",
		Endpoint:        "https://media.example.com",
		AccessKeyID:     "ak",
		AccessKeySecret: "sk",
	})
	if err != nil {
		t.Fatalf("new oss client: %v", err)
	}

	url, err := client.ObjectURL("/covers/1/a.png")
	if err != nil {
		t.Fatalf("object url: %v", err)
	}
	want := "https://media.example.com/covers/1/a.png"
	if url != want {
		t.Fatalf("expected %q, got %q", want, url)
	}
}

// 测试配置public_base_url时优先使用公开访问域名
// 这个场景用于后续接入CDN或自定义域名
func TestObjectURLPrefersPublicBaseURL(t *testing.T) {
	client := newTestClient(t, "https://cdn.example.com/")

	url, err := client.ObjectURL("/avatars/1/a.png")
	if err != nil {
		t.Fatalf("object url: %v", err)
	}
	want := "https://cdn.example.com/avatars/1/a.png"
	if url != want {
		t.Fatalf("expected %q, got %q", want, url)
	}
}

// 测试空对象key会返回错误
// OSS对象必须有明确key 空key不能生成URL
func TestObjectURLRejectsEmptyKey(t *testing.T) {
	client := newTestClient(t, "")

	if _, err := client.ObjectURL("   "); err == nil {
		t.Fatal("expected empty object key to return error")
	}
}

// 测试未初始化的客户端调用方法时会返回错误
// 这些分支保证业务层拿到nil client时不会误返回无效URL
func TestNilClientMethodsReturnError(t *testing.T) {
	var client *Client

	if _, err := client.ObjectURL("videos/a.mp4"); err == nil {
		t.Fatal("expected nil client ObjectURL to return error")
	}
	if err := client.DeleteObject(context.Background(), "videos/a.mp4"); err == nil {
		t.Fatal("expected nil client DeleteObject to return error")
	}
	if _, err := client.UploadFile(context.Background(), "videos/a.mp4", "/tmp/a.mp4"); err == nil {
		t.Fatal("expected nil client UploadFile to return error")
	}
}

// 测试UploadFile在本地文件路径为空或不存在时直接失败
// 这里不测试真实上传 只覆盖上传前的参数校验
func TestUploadFileRejectsInvalidFilePath(t *testing.T) {
	client := newTestClient(t, "")

	if _, err := client.UploadFile(context.Background(), "videos/a.mp4", "  "); err == nil {
		t.Fatal("expected empty upload file path to return error")
	}
	if _, err := client.UploadFile(context.Background(), "videos/a.mp4", t.TempDir()+"/missing.mp4"); err == nil {
		t.Fatal("expected missing upload file to return error")
	}
}

func newTestClient(t *testing.T, publicBaseURL string) *Client {
	t.Helper()

	client, err := NewClient(&config.OSSConfig{
		Bucket:          "examplebucket",
		Region:          "cn-beijing",
		AccessKeyID:     "ak",
		AccessKeySecret: "sk",
		PublicBaseURL:   publicBaseURL,
	})
	if err != nil {
		t.Fatalf("new oss client: %v", err)
	}
	return client
}
