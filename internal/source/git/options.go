package git

import "time"

// Option, GitSource yapılandırma seçeneği.
type Option func(*GitSource)

// WithSince, belirtilen tarihten sonraki commit'leri tarar.
func WithSince(t time.Time) Option {
	return func(s *GitSource) {
		s.since = &t
	}
}

// WithSinceCommit, belirtilen commit'ten HEAD'e kadar olan değişiklikleri tarar.
func WithSinceCommit(hash string) Option {
	return func(s *GitSource) {
		s.sinceCommit = hash
	}
}

// WithBranch, belirtilen branch'ı tarar.
func WithBranch(branch string) Option {
	return func(s *GitSource) {
		s.branch = branch
	}
}

// WithDepth, klonlama derinliğini ayarlar (sadece uzak depolar için).
func WithDepth(depth int) Option {
	return func(s *GitSource) {
		s.depth = depth
	}
}

// WithMaxFileSize, maksimum dosya boyutunu ayarlar.
func WithMaxFileSize(size int64) Option {
	return func(s *GitSource) {
		s.maxFileSize = size
	}
}

// WithBufferSize, chunk kanalı tampon boyutunu ayarlar.
func WithBufferSize(size int) Option {
	return func(s *GitSource) {
		s.bufferSize = size
	}
}
