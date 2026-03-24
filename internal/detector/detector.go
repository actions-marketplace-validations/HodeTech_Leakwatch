// Package detector, sır tespiti için dedektör arayüzlerini tanımlar.
package detector

import (
	"context"

	"github.com/cemililik/leakwatch/pkg/finding"
)

// Detector, belirli bir sır türünü tespit eden bileşeni temsil eder.
type Detector interface {
	// ID, dedektörün benzersiz tanımlayıcısını döndürür (örn: "aws-access-key-id").
	ID() string

	// Description, dedektörün insan tarafından okunabilir açıklamasını döndürür.
	Description() string

	// Keywords, Aho-Corasick ön-filtreleme için anahtar kelimeleri döndürür.
	// Boş döndürürse, ön-filtreleme atlanır ve her chunk'a regex uygulanır.
	Keywords() []string

	// Scan, verilen veriyi tarar ve bulunan potansiyel sırları döndürür.
	Scan(ctx context.Context, data []byte) []RawFinding

	// Severity, bu dedektörün bulguları için varsayılan önem derecesini döndürür.
	Severity() finding.Severity
}

// RawFinding, doğrulanmamış bir ham bulguyu temsil eder.
type RawFinding struct {
	DetectorID string
	Raw        []byte
	RawV2      []byte
	Redacted   string
	ExtraData  map[string]string
}
