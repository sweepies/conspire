package s3util

import (
	"context"
	"io"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

// Helper is a helper object for S3 actions using the AWS SDK
type Helper struct {
	AwsSession *session.Session
	Service    *s3.S3
	Downloader *s3manager.Downloader
	Uploader   *s3manager.Uploader
}

// New creates a new Helper object
func New() *Helper {
	session := connect()
	client := s3.New(session)

	util := Helper{
		AwsSession: session,
		Service:    client,
		Downloader: s3manager.NewDownloaderWithClient(client),
		Uploader:   s3manager.NewUploaderWithClient(client),
	}

	return &util
}

// HeadObject abstracts the S3 HeadObject action
func (h Helper) HeadObject(bucket string, key string) (*s3.HeadObjectOutput, error) {
	input := s3.HeadObjectInput{
		Bucket: &bucket,
		Key:    &key,
	}

	return h.Service.HeadObject(&input)
}

// GetObject abstracts the S3 GetObject action
func (h Helper) GetObject(bucket string, key string) (*s3.GetObjectOutput, error) {
	input := s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &key,
	}

	return h.Service.GetObject(&input)
}

// DeleteObject abstracts the S3 DeleteObject action
func (h Helper) DeleteObject(bucket string, key string) (*s3.DeleteObjectOutput, error) {
	input := s3.DeleteObjectInput{
		Bucket: &bucket,
		Key:    &key,
	}

	return h.Service.DeleteObject(&input)
}

// DownloadObject abstracts an S3 multipart file download
func (h Helper) DownloadObject(bucket string, key string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	input := s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &key,
	}

	buf := aws.NewWriteAtBuffer([]byte{})

	_, err := h.Downloader.DownloadWithContext(ctx, buf, &input)

	bytes := buf.Bytes()

	return bytes, err
}

// UploadObject abstracts an S3 multipart file upload
func (h Helper) UploadObject(bucket string, key string, body io.Reader, contentType string, cacheControl string) (*s3manager.UploadOutput, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	input := s3manager.UploadInput{
		Bucket:       &bucket,
		Key:          &key,
		Body:         body,
		ContentType:  &contentType,
		CacheControl: &cacheControl,
	}

	return h.Uploader.UploadWithContext(ctx, &input)
}

func connect() *session.Session {
	var resolver endpoints.ResolverFunc = func(service, region string, opts ...func(*endpoints.Options)) (endpoints.ResolvedEndpoint, error) {
		return endpoints.ResolvedEndpoint{
			URL: endpoints.AddScheme(viper.GetString("s3_endpoint"), false),
		}, nil
	}

	awsSession, err := session.NewSession(&aws.Config{
		Region:           aws.String(viper.GetString("s3_region")),
		EndpointResolver: resolver,
	})

	if err != nil {
		log.Fatal().Err(err).Send()
	}

	return awsSession
}
