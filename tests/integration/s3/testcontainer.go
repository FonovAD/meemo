package s3

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	tcminio "github.com/testcontainers/testcontainers-go/modules/minio"
)

const (
	testAccessKey = "minioadmin"
	testSecretKey = "minioadmin"
	testRegion    = "us-east-1"
)

type TestMinioContainer struct {
	Container  *tcminio.MinioContainer
	Client     *s3.Client
	BucketName string
	Endpoint   string
}

func SetupMinioContainer(t *testing.T) (*TestMinioContainer, func()) {
	ctx := context.Background()

	minioContainer, err := tcminio.Run(ctx,
		"minio/minio:latest",
		tcminio.WithUsername(testAccessKey),
		tcminio.WithPassword(testSecretKey),
	)
	if err != nil {
		t.Fatalf("Failed to start minio container: %v", err)
	}

	endpoint, err := minioContainer.ConnectionString(ctx)
	if err != nil {
		t.Fatalf("Failed to get minio endpoint: %v", err)
	}
	//nolint:staticcheck
	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL:               "http://" + endpoint,
			HostnameImmutable: true,
		}, nil
	})

	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(testRegion),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(testAccessKey, testSecretKey, "")),
		config.WithEndpointResolverWithOptions(customResolver), //nolint:staticcheck
	)
	if err != nil {
		t.Fatalf("Failed to load AWS config: %v", err)
	}

	s3Client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = true
	})

	testBucket := fmt.Sprintf("test-bucket-%d", time.Now().UnixNano())
	_, err = s3Client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: aws.String(testBucket),
	})
	if err != nil {
		t.Fatalf("Failed to create test bucket: %v", err)
	}

	tc := &TestMinioContainer{
		Container:  minioContainer,
		Client:     s3Client,
		BucketName: testBucket,
		Endpoint:   endpoint,
	}

	cleanup := func() {
		cleanupBucket(t, s3Client, testBucket)
		if err := minioContainer.Terminate(ctx); err != nil {
			t.Logf("Warning: failed to terminate minio container: %v", err)
		}
	}

	return tc, cleanup
}

func cleanupBucket(t *testing.T, client *s3.Client, bucketName string) {
	ctx := context.Background()

	listOutput, err := client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		t.Logf("Error listing objects: %v", err)
		return
	}

	if len(listOutput.Contents) > 0 {
		var objectIds []types.ObjectIdentifier
		for _, obj := range listOutput.Contents {
			objectIds = append(objectIds, types.ObjectIdentifier{
				Key: obj.Key,
			})
		}

		_, err = client.DeleteObjects(ctx, &s3.DeleteObjectsInput{
			Bucket: aws.String(bucketName),
			Delete: &types.Delete{
				Objects: objectIds,
				Quiet:   aws.Bool(true),
			},
		})
		if err != nil {
			t.Logf("Error deleting objects: %v", err)
		}
	}

	_, err = client.DeleteBucket(ctx, &s3.DeleteBucketInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		t.Logf("Failed to remove bucket %s: %v", bucketName, err)
	}
}

func SetupS3ClientForTest(t *testing.T) (*s3.Client, string, func()) {
	tc, cleanup := SetupMinioContainer(t)
	return tc.Client, tc.BucketName, cleanup
}
