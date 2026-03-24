package filesystem

// Option, FilesystemSource yapılandırma seçeneği.
type Option func(*FilesystemSource)

// WithMaxFileSize, maksimum dosya boyutunu ayarlar.
func WithMaxFileSize(size int64) Option {
	return func(s *FilesystemSource) {
		s.maxFileSize = size
	}
}

// WithExcludeExtensions, hariç tutulacak dosya uzantılarını ayarlar.
func WithExcludeExtensions(exts []string) Option {
	return func(s *FilesystemSource) {
		s.excludeExts = exts
	}
}

// WithExcludePaths, hariç tutulacak yol desenlerini ayarlar.
func WithExcludePaths(paths []string) Option {
	return func(s *FilesystemSource) {
		s.excludePaths = paths
	}
}

// WithBufferSize, chunk kanalı tampon boyutunu ayarlar.
func WithBufferSize(size int) Option {
	return func(s *FilesystemSource) {
		s.bufferSize = size
	}
}
