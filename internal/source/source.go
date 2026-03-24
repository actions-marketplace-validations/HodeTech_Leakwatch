// Package source, tarama kaynağı arayüzlerini tanımlar.
package source

import (
	"context"

	"github.com/cemililik/leakwatch/pkg/finding"
)

// Source, taranacak veri kaynağını temsil eder.
// Her kaynak türü (Git, dosya sistemi, container) bu arayüzü uygular.
type Source interface {
	// Type, kaynağın türünü döndürür (örn: "git", "filesystem", "container").
	Type() string

	// Chunks, taranacak veri parçalarını bir kanal üzerinden gönderir.
	// Context iptal edildiğinde kanal kapatılır.
	Chunks(ctx context.Context) <-chan Chunk

	// Validate, kaynağın erişilebilir ve geçerli olduğunu kontrol eder.
	Validate() error
}

// Chunk, taranacak en küçük veri birimini temsil eder.
type Chunk struct {
	// Data, taranacak ham içerik.
	Data []byte

	// SourceMetadata, bulgunun nereden geldiğini tanımlayan bağlam bilgisi.
	SourceMetadata finding.SourceMetadata
}
