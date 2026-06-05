package storage

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

const PresignTTL = 15 * time.Minute

type Config struct {
	Endpoint      string
	AccessKey     string
	SecretKey     string
	Bucket        string
	Region        string
	UseSSL        bool
	PublicBaseURL string
}

type Client struct {
	minio      *minio.Client
	bucket     string
	publicBase string
}

type ObjectInfo struct {
	Size        int64
	ContentType string
}

func New(ctx context.Context, cfg Config) (*Client, error) {
	if cfg.Endpoint == "" {
		return nil, fmt.Errorf("S3 endpoint is required")
	}
	if cfg.Bucket == "" {
		return nil, fmt.Errorf("S3 bucket is required")
	}
	if cfg.PublicBaseURL == "" {
		return nil, fmt.Errorf("upload public base URL is required")
	}
	region := cfg.Region
	if region == "" {
		region = "us-east-1"
	}

	endpoint := strings.TrimPrefix(strings.TrimPrefix(cfg.Endpoint, "https://"), "http://")

	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
		Region: region,
	})
	if err != nil {
		return nil, err
	}

	c := &Client{
		minio:      client,
		bucket:     cfg.Bucket,
		publicBase: strings.TrimRight(cfg.PublicBaseURL, "/"),
	}

	if err := c.ensureBucket(ctx); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *Client) ensureBucket(ctx context.Context) error {
	exists, err := c.minio.BucketExists(ctx, c.bucket)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	return c.minio.MakeBucket(ctx, c.bucket, minio.MakeBucketOptions{})
}

func (c *Client) PublicURL(objectKey string) string {
	return c.publicBase + "/" + strings.TrimPrefix(objectKey, "/")
}

func (c *Client) PresignPut(ctx context.Context, objectKey, contentType string) (string, error) {
	url, err := c.minio.PresignedPutObject(ctx, c.bucket, objectKey, PresignTTL)
	if err != nil {
		return "", err
	}
	_ = contentType
	return url.String(), nil
}

func (c *Client) StatObject(ctx context.Context, objectKey string) (*ObjectInfo, error) {
	info, err := c.minio.StatObject(ctx, c.bucket, objectKey, minio.StatObjectOptions{})
	if err != nil {
		return nil, err
	}
	return &ObjectInfo{
		Size:        info.Size,
		ContentType: info.ContentType,
	}, nil
}

func (c *Client) ReadHeader(ctx context.Context, objectKey string, n int) ([]byte, error) {
	if n <= 0 {
		n = 512
	}
	obj, err := c.minio.GetObject(ctx, c.bucket, objectKey, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	defer obj.Close()

	buf := make([]byte, n)
	read, err := io.ReadFull(obj, buf)
	if err != nil && err != io.ErrUnexpectedEOF && err != io.EOF {
		return nil, err
	}
	return buf[:read], nil
}

func ExtensionForContentType(contentType string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(contentType)) {
	case "image/jpeg":
		return ".jpg", nil
	case "image/png":
		return ".png", nil
	case "image/webp":
		return ".webp", nil
	default:
		return "", fmt.Errorf("unsupported content type %q", contentType)
	}
}

func ValidateImageHeader(contentType string, header []byte) error {
	if len(header) < 3 {
		return fmt.Errorf("file is too small to be a valid image")
	}

	switch strings.ToLower(contentType) {
	case "image/jpeg":
		if header[0] != 0xFF || header[1] != 0xD8 || header[2] != 0xFF {
			return fmt.Errorf("file is not a valid JPEG")
		}
	case "image/png":
		if len(header) < 8 || string(header[:8]) != "\x89PNG\r\n\x1a\n" {
			return fmt.Errorf("file is not a valid PNG")
		}
	case "image/webp":
		if len(header) < 12 || string(header[:4]) != "RIFF" || string(header[8:12]) != "WEBP" {
			return fmt.Errorf("file is not a valid WebP")
		}
	default:
		return fmt.Errorf("unsupported content type %q", contentType)
	}
	return nil
}

func ManagedURL(publicBaseURL, rawURL string) bool {
	base := strings.TrimRight(publicBaseURL, "/")
	return strings.HasPrefix(strings.TrimSpace(rawURL), base+"/")
}

func ParsePublicBaseHost(publicBaseURL string) (string, error) {
	u, err := url.Parse(publicBaseURL)
	if err != nil {
		return "", err
	}
	if u.Host == "" {
		return "", fmt.Errorf("invalid public base URL")
	}
	return u.Host, nil
}
