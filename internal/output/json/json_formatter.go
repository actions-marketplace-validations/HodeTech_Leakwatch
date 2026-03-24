// Package json, JSON çıktı formatlayıcısını sağlar.
package json

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/cemililik/leakwatch/pkg/finding"
)

// Formatter, bulguları JSON formatında çıktılar.
type Formatter struct {
	// ShowRaw, true olduğunda bulguların Raw alanını çıktıya dahil eder.
	// false olduğunda Raw alanı güvenlik amacıyla çıktıdan çıkarılır (defense in depth).
	ShowRaw bool
}

// Format, bulguları JSON formatında writer'a yazar.
// ShowRaw false ise Raw alanları aktif olarak temizlenir.
func (f *Formatter) Format(w io.Writer, findings []finding.Finding) error {
	output := make([]finding.Finding, len(findings))
	copy(output, findings)

	if !f.ShowRaw {
		for i := range output {
			output[i].Raw = ""
		}
	}

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(output); err != nil {
		return fmt.Errorf("JSON çıktı yazılamadı: %w", err)
	}
	return nil
}

// FileExtension, JSON dosya uzantısını döndürür.
func (f *Formatter) FileExtension() string {
	return ".json"
}
