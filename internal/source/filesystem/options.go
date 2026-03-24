package filesystem

// Option configures a FilesystemSource.
type Option func(*FilesystemSource)

// WithMaxFileSize sets the maximum file size to scan.
func WithMaxFileSize(size int64) Option {
	return func(s *FilesystemSource) {
		s.maxFileSize = size
	}
}

// WithExcludeExtensions sets file extensions to exclude from scanning.
func WithExcludeExtensions(exts []string) Option {
	return func(s *FilesystemSource) {
		s.excludeExts = exts
	}
}

// WithExcludePaths sets path patterns to exclude from scanning.
func WithExcludePaths(paths []string) Option {
	return func(s *FilesystemSource) {
		s.excludePaths = paths
	}
}

// WithBufferSize sets the chunk channel buffer size.
func WithBufferSize(size int) Option {
	return func(s *FilesystemSource) {
		s.bufferSize = size
	}
}
