package container

// Option configures a ContainerSource.
type Option func(*ContainerSource)

// WithMaxFileSize sets the maximum file size to extract from layers.
// Values less than or equal to zero are ignored.
func WithMaxFileSize(size int64) Option {
	return func(s *ContainerSource) {
		if size <= 0 {
			return
		}
		s.maxFileSize = size
	}
}

// WithBufferSize sets the chunk channel buffer size.
// Values less than or equal to zero are ignored.
func WithBufferSize(size int) Option {
	return func(s *ContainerSource) {
		if size <= 0 {
			return
		}
		s.bufferSize = size
	}
}
