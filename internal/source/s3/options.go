// Package s3 provides an AWS S3 bucket scan source.
package s3

// Option configures an S3Source.
type Option func(*S3Source)

// WithPrefix limits the scan to objects matching the given key prefix.
func WithPrefix(prefix string) Option {
	return func(s *S3Source) {
		s.prefix = prefix
	}
}

// WithMaxFileSize sets the maximum object size to download and scan.
// Objects larger than this value are skipped.
func WithMaxFileSize(size int64) Option {
	return func(s *S3Source) {
		s.maxFileSize = size
	}
}

// WithBufferSize sets the channel buffer size for the chunk channel.
func WithBufferSize(size int) Option {
	return func(s *S3Source) {
		if size > 0 {
			s.bufferSize = size
		}
	}
}

// WithRegion sets the AWS region for the S3 client.
func WithRegion(region string) Option {
	return func(s *S3Source) {
		s.region = region
	}
}
