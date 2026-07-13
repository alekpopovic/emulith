package manifest

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
	awss3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	awssqs "github.com/aws/aws-sdk-go-v2/service/sqs"
	"gopkg.in/yaml.v3"
)

type Manifest struct {
	Version   int        `yaml:"version"`
	Project   string     `yaml:"project"`
	Resources []Resource `yaml:"resources"`
}
type Resource struct {
	Type     string            `yaml:"type"`
	Name     string            `yaml:"name"`
	Region   string            `yaml:"region,omitempty"`
	Metadata map[string]string `yaml:"metadata,omitempty"`
}

var projectPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_-]{0,62}$`)
var bucketPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9.-]{1,61}[a-z0-9]$`)
var queuePattern = regexp.MustCompile(`^[A-Za-z0-9_-]{1,80}$`)

func Parse(data []byte) (Manifest, error) {
	var m Manifest
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	if err := decoder.Decode(&m); err != nil {
		return m, fmt.Errorf("decode manifest: %w", err)
	}
	if err := m.Validate(); err != nil {
		return m, err
	}
	return m, nil
}
func (m Manifest) Validate() error {
	if m.Version != 1 {
		return fmt.Errorf("version: expected 1")
	}
	if !projectPattern.MatchString(m.Project) {
		return fmt.Errorf("project: must match %s", projectPattern)
	}
	seen := map[string]bool{}
	for i, r := range m.Resources {
		prefix := fmt.Sprintf("resources[%d]", i)
		if r.Type == "" {
			return fmt.Errorf("%s.type: required", prefix)
		}
		if r.Name == "" {
			return fmt.Errorf("%s.name: required", prefix)
		}
		id := r.Type + "\x00" + r.Name
		if seen[id] {
			return fmt.Errorf("%s: duplicate resource", prefix)
		}
		seen[id] = true
		switch r.Type {
		case "aws.s3.bucket":
			if !bucketPattern.MatchString(r.Name) || strings.Contains(r.Name, "..") {
				return fmt.Errorf("%s.name: invalid bucket name", prefix)
			}
		case "aws.sqs.queue":
			if strings.HasSuffix(r.Name, ".fifo") {
				return fmt.Errorf("%s.name: FIFO queues unsupported", prefix)
			}
		case "azure.storage.container":
			if len(r.Name) < 3 || len(r.Name) > 63 || !regexp.MustCompile(`^[a-z0-9-]+$`).MatchString(r.Name) {
				return fmt.Errorf("%s.name: invalid Azure container name", prefix)
			}
		case "azure.storage.queue":
			if !validAzureQueue(r.Name) {
				return fmt.Errorf("%s.name: invalid Azure queue name", prefix)
			}
		case "azure.storage.table":
			if len(r.Name) < 3 || len(r.Name) > 63 || !regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_]*$`).MatchString(r.Name) {
				return fmt.Errorf("%s.name: invalid Azure table name", prefix)
			}
			if !queuePattern.MatchString(r.Name) {
				return fmt.Errorf("%s.name: invalid queue name", prefix)
			}
			if r.Region != "" {
				return fmt.Errorf("%s.region: only valid for aws.s3.bucket", prefix)
			}
		default:
			return fmt.Errorf("%s.type: unsupported resource type %q", prefix, r.Type)
		}
	}
	return nil
}
func validAzureQueue(n string) bool {
	if len(n) < 3 || len(n) > 63 || n[0] == '-' || n[len(n)-1] == '-' || strings.Contains(n, "--") {
		return false
	}
	for _, c := range n {
		if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-') {
			return false
		}
	}
	return true
}
func ValidateEndpoint(raw string) error {
	u, err := url.Parse(raw)
	if err != nil || u.Scheme == "" || u.Hostname() == "" {
		return fmt.Errorf("invalid endpoint %q", raw)
	}
	host := u.Hostname()
	ip := net.ParseIP(host)
	if host != "localhost" && (ip == nil || !ip.IsLoopback()) {
		return fmt.Errorf("endpoint host %q is not loopback", host)
	}
	return nil
}

type Applier struct {
	S3  *awss3.Client
	SQS *awssqs.Client
}

func NewApplier(endpoint string, client *http.Client) (*Applier, error) {
	if err := ValidateEndpoint(endpoint); err != nil {
		return nil, err
	}
	cfg := aws.Config{Region: "us-east-1", Credentials: credentials.NewStaticCredentialsProvider("test", "test", ""), HTTPClient: client}
	return &Applier{S3: awss3.NewFromConfig(cfg, func(o *awss3.Options) { o.BaseEndpoint = aws.String(endpoint); o.UsePathStyle = true }), SQS: awssqs.NewFromConfig(cfg, func(o *awssqs.Options) { o.BaseEndpoint = aws.String(endpoint) })}, nil
}
func (a *Applier) Apply(ctx context.Context, m Manifest, out io.Writer) error {
	for _, r := range m.Resources {
		switch r.Type {
		case "aws.s3.bucket":
			listed, err := a.S3.ListBuckets(ctx, &awss3.ListBucketsInput{})
			if err != nil {
				return fmt.Errorf("apply %s: %w", r.Name, err)
			}
			exists := false
			for _, b := range listed.Buckets {
				if aws.ToString(b.Name) == r.Name {
					exists = true
				}
			}
			if !exists {
				input := &awss3.CreateBucketInput{Bucket: aws.String(r.Name)}
				if r.Region != "" && r.Region != "us-east-1" {
					input.CreateBucketConfiguration = &awss3types.CreateBucketConfiguration{LocationConstraint: awss3types.BucketLocationConstraint(r.Region)}
				}
				if _, err := a.S3.CreateBucket(ctx, input); err != nil {
					return fmt.Errorf("apply %s: %w", r.Name, err)
				}
			}
			fmt.Fprintf(out, "applied %s %s\n", r.Type, r.Name)
		case "aws.sqs.queue":
			if _, err := a.SQS.CreateQueue(ctx, &awssqs.CreateQueueInput{QueueName: aws.String(r.Name)}); err != nil {
				return fmt.Errorf("apply %s: %w", r.Name, err)
			}
			fmt.Fprintf(out, "applied %s %s\n", r.Type, r.Name)
		}
	}
	fmt.Fprintf(out, "applied %d resources for project %s\n", len(m.Resources), m.Project)
	return nil
}
