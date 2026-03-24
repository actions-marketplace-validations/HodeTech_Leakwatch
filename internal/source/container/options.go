package container

// Option configures a ContainerSource.
type Option func(*ContainerSource)

// WithMaxFileSize sets the maximum file size to extract from layers.
func WithMaxFileSize(size int64) Option {
	return func(s *ContainerSource) {
		s.maxFileSize = size
	}
}

// WithBufferSize sets the chunk channel buffer size.
func WithBufferSize(size int) Option {
	return func(s *ContainerSource) {
		s.bufferSize = size
	}
}
