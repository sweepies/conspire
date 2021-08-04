package util

import (
	"context"
	"errors"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/rs/zerolog/log"
)

type S3 struct {
	Client     *s3.Client
	Downloader *manager.Downloader
	Uploader   *manager.Uploader
}

// NewS3 creates a new S3 object
func NewS3(c chan *S3, endpoint string) {
	resolver := aws.EndpointResolverFunc(func(service, region string) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL: endpoint,
		}, nil
	})

	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithEndpointResolver(resolver))

	if err != nil {
		log.Fatal().Err(err).Msg("Error loading AWS configuration")
	}

	client := s3.NewFromConfig(cfg)

	c <- &S3{
		Client:     client,
		Downloader: manager.NewDownloader(client),
		Uploader:   manager.NewUploader(client),
	}
}

func (s S3) ObjectExists(bucket, key string) (bool, error) {
	_, err := s.HeadObject(bucket, key)

	if err != nil {
		var nsk *types.NoSuchKey
		if errors.As(err, &nsk) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

// HeadObject abstracts the S3 HeadObject action
func (s S3) HeadObject(bucket, key string) (*s3.HeadObjectOutput, error) {
	input := s3.HeadObjectInput{
		Bucket: &bucket,
		Key:    &key,
	}

	return s.Client.HeadObject(context.TODO(), &input)
}

// GetObject abstracts the S3 GetObject action
func (s S3) GetObject(bucket, key string) (*s3.GetObjectOutput, error) {
	input := s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &key,
	}

	return s.Client.GetObject(context.TODO(), &input)
}

// DeleteObject abstracts the S3 DeleteObject action
func (s S3) DeleteObject(bucket, key string) (*s3.DeleteObjectOutput, error) {
	input := s3.DeleteObjectInput{
		Bucket: &bucket,
		Key:    &key,
	}

	return s.Client.DeleteObject(context.TODO(), &input)
}

// DownloadObject abstracts an S3 multipart file download
func (s S3) DownloadObject(bucket, key string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	input := s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &key,
	}

	buf := manager.NewWriteAtBuffer([]byte{})

	_, err := s.Downloader.Download(ctx, buf, &input)

	bytes := buf.Bytes()

	return bytes, err
}

// UploadObject abstracts an S3 multipart file upload
func (s S3) UploadObject(bucket, key string, body io.Reader, contentType, cacheControl string, makePublic bool) (*manager.UploadOutput, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	input := s3.PutObjectInput{
		Bucket:       &bucket,
		Key:          &key,
		Body:         body,
		ContentType:  &contentType,
		CacheControl: &cacheControl,
	}

	if makePublic {
		input.ACL = types.ObjectCannedACLPublicRead
	}

	return s.Uploader.Upload(ctx, &input)
}
