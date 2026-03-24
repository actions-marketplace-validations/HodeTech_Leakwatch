// Package finding, Leakwatch bulgu modelini tanımlar.
// Bu paket dışa açıktır ve CI eklentileri gibi harici tüketiciler tarafından kullanılabilir.
package finding

import (
	"encoding/json"
	"fmt"
	"time"
)

// Severity, bulgu önem derecesi.
type Severity int

const (
	SeverityLow      Severity = iota // Düşük
	SeverityMedium                   // Orta
	SeverityHigh                     // Yüksek
	SeverityCritical                 // Kritik
)

// severityToString, Severity değerlerini string'e eşler.
var severityToString = map[Severity]string{
	SeverityLow:      "low",
	SeverityMedium:   "medium",
	SeverityHigh:     "high",
	SeverityCritical: "critical",
}

// stringToSeverity, string değerlerini Severity'ye eşler.
var stringToSeverity = map[string]Severity{
	"low":      SeverityLow,
	"medium":   SeverityMedium,
	"high":     SeverityHigh,
	"critical": SeverityCritical,
}

// String, Severity'nin okunabilir gösterimini döndürür.
func (s Severity) String() string {
	if str, ok := severityToString[s]; ok {
		return str
	}
	return "unknown"
}

// MarshalJSON, Severity'yi JSON string olarak serileştirir.
func (s Severity) MarshalJSON() ([]byte, error) {
	str := s.String()
	if str == "unknown" {
		return nil, fmt.Errorf("geçersiz Severity değeri: %d", int(s))
	}
	return json.Marshal(str)
}

// UnmarshalJSON, JSON string'den Severity değeri çözümler.
func (s *Severity) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return fmt.Errorf("Severity JSON çözümlenemedi: %w", err)
	}
	val, ok := stringToSeverity[str]
	if !ok {
		return fmt.Errorf("geçersiz Severity değeri: %q", str)
	}
	*s = val
	return nil
}

// VerificationStatus, doğrulama durumu.
type VerificationStatus int

const (
	StatusUnverified      VerificationStatus = iota // Doğrulama yapılmadı
	StatusVerifiedActive                             // Doğrulandı: sır aktif
	StatusVerifiedInactive                           // Doğrulandı: sır devre dışı
	StatusVerifyError                                // Doğrulama sırasında hata
)

// verificationStatusToString, VerificationStatus değerlerini string'e eşler.
var verificationStatusToString = map[VerificationStatus]string{
	StatusUnverified:      "unverified",
	StatusVerifiedActive:  "verified_active",
	StatusVerifiedInactive: "verified_inactive",
	StatusVerifyError:     "verify_error",
}

// stringToVerificationStatus, string değerlerini VerificationStatus'a eşler.
var stringToVerificationStatus = map[string]VerificationStatus{
	"unverified":        StatusUnverified,
	"verified_active":   StatusVerifiedActive,
	"verified_inactive": StatusVerifiedInactive,
	"verify_error":      StatusVerifyError,
}

// String, VerificationStatus'un okunabilir gösterimini döndürür.
func (v VerificationStatus) String() string {
	if str, ok := verificationStatusToString[v]; ok {
		return str
	}
	return "unknown"
}

// MarshalJSON, VerificationStatus'u JSON string olarak serileştirir.
func (v VerificationStatus) MarshalJSON() ([]byte, error) {
	str := v.String()
	if str == "unknown" {
		return nil, fmt.Errorf("geçersiz VerificationStatus değeri: %d", int(v))
	}
	return json.Marshal(str)
}

// UnmarshalJSON, JSON string'den VerificationStatus değeri çözümler.
func (v *VerificationStatus) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return fmt.Errorf("VerificationStatus JSON çözümlenemedi: %w", err)
	}
	val, ok := stringToVerificationStatus[str]
	if !ok {
		return fmt.Errorf("geçersiz VerificationStatus değeri: %q", str)
	}
	*v = val
	return nil
}

// VerificationResult, doğrulama sonucunu temsil eder.
type VerificationResult struct {
	Status    VerificationStatus `json:"status"`
	Message   string             `json:"message,omitempty"`
	ExtraData map[string]string  `json:"extra_data,omitempty"`
}

// SourceMetadata, bir bulgunun kaynak bilgisini tanımlar.
type SourceMetadata struct {
	SourceType string `json:"source_type"`

	// Git'e özgü
	Repository string    `json:"repository,omitempty"`
	Commit     string    `json:"commit,omitempty"`
	Author     string    `json:"author,omitempty"`
	Email      string    `json:"email,omitempty"`
	Date       time.Time `json:"date,omitempty"`
	Branch     string    `json:"branch,omitempty"`

	// Dosyaya özgü
	FilePath string `json:"file_path,omitempty"`
	Line     int    `json:"line,omitempty"`

	// Container'a özgü
	Image    string `json:"image,omitempty"`
	Layer    string `json:"layer,omitempty"`
	LayerIdx int    `json:"layer_idx,omitempty"`
}

// Finding, tam olarak zenginleştirilmiş bir bulguyu temsil eder.
type Finding struct {
	ID             string             `json:"id"`
	DetectorID     string             `json:"detector_id"`
	Severity       Severity           `json:"severity"`
	Raw            string             `json:"raw,omitempty"`
	Redacted       string             `json:"redacted"`
	SourceMetadata SourceMetadata     `json:"source"`
	Verification   VerificationResult `json:"verification"`
	DetectedAt     time.Time          `json:"detected_at"`
	Entropy        float64            `json:"entropy,omitempty"`
	ExtraData      map[string]string  `json:"extra_data,omitempty"`
}
