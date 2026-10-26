package storage

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"io"
	"net/url"
	"os"
)

type Media struct {
	Client            *minio.Client
	Bucket, PublicURL string
}

func NewMedia() (*Media, error) {
	c, e := minio.New(os.Getenv("MINIO_ENDPOINT"), &minio.Options{Creds: credentials.NewStaticV4(os.Getenv("MINIO_ACCESS_KEY"), os.Getenv("MINIO_SECRET_KEY"), ""), Secure: false})
	return &Media{c, "cars", os.Getenv("MEDIA_PUBLIC_URL")}, e
}
func (m *Media) URL(name string) string {
	if name == "" {
		return ""
	}
	return m.PublicURL + "/" + url.PathEscape(name)
}
func (m *Media) Upload(ctx context.Context, r io.Reader, size int64, ext, contentType string) (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	name := hex.EncodeToString(b) + ext
	_, err := m.Client.PutObject(ctx, m.Bucket, name, r, size, minio.PutObjectOptions{ContentType: contentType})
	return name, err
}
func (m *Media) Remove(ctx context.Context, name string) {
	if name != "" {
		_ = m.Client.RemoveObject(ctx, m.Bucket, name, minio.RemoveObjectOptions{})
	}
}
