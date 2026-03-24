// Package output, çıktı formatlayıcı arayüzlerini tanımlar.
package output

import (
	"io"

	"github.com/cemililik/leakwatch/pkg/finding"
)

// Formatter, bulguları belirli bir formatta çıktılayan bileşeni temsil eder.
type Formatter interface {
	// Format, bulguları belirtilen writer'a yazar.
	Format(w io.Writer, findings []finding.Finding) error

	// FileExtension, bu formatın dosya uzantısını döndürür.
	FileExtension() string
}
