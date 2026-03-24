# Leakwatch - Geliştirme Standartları ve Altyapı

> **Belge Versiyonu:** 1.0
> **Tarih:** 2026-03-24
> **Durum:** Taslak

---

## 1. Geliştirme Ortamı Gereksinimleri

### 1.1 Zorunlu Araçlar

| Araç | Minimum Versiyon | Amaç |
|------|------------------|------|
| Go | 1.22+ | Ana programlama dili |
| Git | 2.30+ | Versiyon kontrol |
| golangci-lint | 1.57+ | Statik analiz ve linting |
| goreleaser | 2.0+ | Build ve release otomasyonu |
| pre-commit | 3.0+ | Git hook yönetimi |

### 1.2 Opsiyonel Araçlar

| Araç | Amaç |
|------|------|
| Docker | Container imaj testleri |
| cobra-cli | CLI iskelet oluşturma |
| govulncheck | Güvenlik açığı taraması |
| gofumpt | Strict kod formatlama |
| delve (dlv) | Hata ayıklama |

### 1.3 IDE Desteği

- **VS Code:** Go eklentisi (Go Team at Google)
- **GoLand:** JetBrains (birinci sınıf Go desteği)
- **Vim/Neovim:** gopls LSP

---

## 2. Kod Standartları

### 2.1 Go Kodlama Kuralları

Leakwatch, aşağıdaki stil rehberlerini takip eder:

1. **Effective Go** — Go ekibinin resmi stil rehberi
2. **Go Code Review Comments** — Yaygın inceleme notları
3. **Uber Go Style Guide** — Ek kurumsal standartlar

### 2.2 Adlandırma Kuralları

| Öğe | Kural | Örnek |
|-----|-------|-------|
| Paket | Kısa, küçük harf, tek kelime | `detector`, `engine`, `output` |
| Dışa açık fonksiyon | PascalCase | `ScanRepository()` |
| Dahili fonksiyon | camelCase | `parseConfig()` |
| Arayüz (Interface) | PascalCase, "-er" soneki | `Detector`, `Verifier`, `Formatter` |
| Sabit (Constant) | PascalCase veya SCREAMING_SNAKE | `MaxFileSize`, `StatusVerifiedActive` |
| Değişken | camelCase | `chunkSize`, `workerCount` |
| Dosya adı | snake_case | `aws_access_key.go`, `worker_pool.go` |
| Test dosyası | `_test.go` soneki | `engine_test.go` |

### 2.3 Paket Organizasyon Kuralları

```
internal/   → Dışarıdan erişilemez paketler (uygulama detayları)
pkg/        → Dışarıdan erişilebilir paketler (kütüphane olarak kullanım)
cmd/        → CLI komut tanımları (ince katman, iş mantığı içermez)
```

- `cmd/` paketi sadece CLI flag tanımları ve bağlama (wiring) içerir
- İş mantığı (business logic) `internal/` altındadır
- Dışarıya açılması istenen tipler `pkg/` altındadır

### 2.4 Hata Yönetimi

```go
// ✅ DOĞRU: Hataları sarmalayarak bağlam ekle
if err != nil {
    return fmt.Errorf("git repo açılamadı %s: %w", path, err)
}

// ❌ YANLIŞ: Çıplak hata döndürme
if err != nil {
    return err
}

// ✅ DOĞRU: Sentinel hatalar tanımla
var (
    ErrSourceNotFound   = errors.New("kaynak bulunamadı")
    ErrInvalidConfig    = errors.New("geçersiz yapılandırma")
    ErrVerifyTimeout    = errors.New("doğrulama zaman aşımı")
)

// ✅ DOĞRU: Context iptali kontrol et
select {
case <-ctx.Done():
    return ctx.Err()
default:
}
```

### 2.5 Loglama Standartları

```go
// ✅ DOĞRU: Yapılandırılmış loglama (log/slog)
slog.Info("tarama tamamlandı",
    "source", "git",
    "findings", len(findings),
    "duration", elapsed,
)

// ❌ YANLIŞ: fmt.Println veya log.Printf
fmt.Println("Tarama tamamlandı")
log.Printf("Bulgu: %d", len(findings))

// ❌ YANLIŞ: Sır içeriğini loglama
slog.Info("sır bulundu", "raw", secretValue) // ASLA YAPMA
```

---

## 3. Test Standartları

### 3.1 Test Piramidi

```mermaid
block-beta
    columns 1
    block:e2e["E2E Testler (az sayıda)\nCLI komutlarının uçtan uca testi"]:1
    end
    block:integration["Entegrasyon Testleri (orta)\nGerçek git repo, gerçek dosya sistemi"]:1
    end
    block:unit["Birim Testler (çok sayıda)\nHer fonksiyon, her dedektör, her parser"]:1
    end
```

### 3.2 Test Kapsam Hedefleri

| Paket | Minimum Kapsam |
|-------|---------------|
| `internal/detector/*` | %95 |
| `internal/engine/*` | %85 |
| `internal/source/*` | %80 |
| `internal/verifier/*` | %75 |
| `internal/entropy/*` | %95 |
| `internal/matcher/*` | %90 |
| `internal/output/*` | %85 |
| **Genel Hedef** | **%80+** |

### 3.3 Test Yazım Kuralları

```go
// ✅ DOĞRU: Table-driven testler
func TestAWSAccessKeyDetector(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected int // beklenen bulgu sayısı
    }{
        {
            name:     "geçerli AWS access key",
            input:    "AKIAIOSFODNN7EXAMPLE",
            expected: 1,
        },
        {
            name:     "test/placeholder key",
            input:    "AKIAIOSFODNN7XXXXXXX",
            expected: 1, // Pattern eşleşir, doğrulama ile ayrılır
        },
        {
            name:     "eşleşme yok",
            input:    "bu normal bir metin",
            expected: 0,
        },
    }

    d := &AWSAccessKeyID{}
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            findings := d.Scan(context.Background(), []byte(tt.input))
            assert.Len(t, findings, tt.expected)
        })
    }
}

// ✅ DOĞRU: io/fs ile bellek içi dosya sistemi testi
func TestFilesystemSource(t *testing.T) {
    fsys := fstest.MapFS{
        "config.yaml": &fstest.MapFile{
            Data: []byte("api_key: AKIAIOSFODNN7EXAMPLE"),
        },
        "main.go": &fstest.MapFile{
            Data: []byte("package main"),
        },
    }
    // fsys'i Source'a ver, test et
}
```

### 3.4 Test Adlandırma

```
Test<Fonksiyon>_<Senaryo>_<BeklenenSonuç>

Örnekler:
- TestScanGit_ValidRepo_ReturnsFindings
- TestShannonEntropy_HighEntropyString_AboveThreshold
- TestAWSVerifier_InvalidKey_ReturnsInactive
- TestEngine_CancelledContext_StopsGracefully
```

### 3.5 Mock ve Stub Kullanımı

```go
// Arayüzlere karşı test et, mock'lar doğal olarak gelir
type mockDetector struct {
    id       string
    keywords []string
    findings []RawFinding
}

func (m *mockDetector) ID() string            { return m.id }
func (m *mockDetector) Keywords() []string     { return m.keywords }
func (m *mockDetector) Scan(_ context.Context, _ []byte) []RawFinding {
    return m.findings
}
```

---

## 4. CI/CD Pipeline

### 4.1 CI Workflow (.github/workflows/ci.yml)

```yaml
name: CI

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        go-version: ['1.22', '1.23']
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: ${{ matrix.go-version }}
      - uses: actions/cache@v4
        with:
          path: ~/go/pkg/mod
          key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
      - run: go test -race -coverprofile=coverage.out ./...
      - run: go build ./...

  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.23'
      - uses: golangci/golangci-lint-action@v6
        with:
          version: latest

  security:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.23'
      - run: go install golang.org/x/vuln/cmd/govulncheck@latest
      - run: govulncheck ./...
```

### 4.2 Release Workflow (.github/workflows/release.yml)

```yaml
name: Release

on:
  push:
    tags: ['v*']

permissions:
  contents: write

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0
      - uses: actions/setup-go@v5
        with:
          go-version: '1.23'
      - uses: goreleaser/goreleaser-action@v6
        with:
          version: '~> v2'
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

### 4.3 GoReleaser Yapılandırması (.goreleaser.yml)

```yaml
version: 2

builds:
  - env:
      - CGO_ENABLED=0
    goos:
      - linux
      - darwin
      - windows
    goarch:
      - amd64
      - arm64
    ldflags:
      - -s -w
      - -X main.version={{.Version}}
      - -X main.commit={{.Commit}}
      - -X main.date={{.Date}}

archives:
  - format: tar.gz
    name_template: >-
      {{ .ProjectName }}_{{ .Version }}_{{ .Os }}_{{ .Arch }}
    format_overrides:
      - goos: windows
        format: zip

checksum:
  name_template: 'checksums.txt'

changelog:
  sort: asc
  filters:
    exclude:
      - '^docs:'
      - '^test:'
      - '^ci:'
```

---

## 5. Git Workflow

### 5.1 Branching Stratejisi: GitHub Flow

```mermaid
gitgraph
    commit id: "init"
    branch feature/scan-git
    commit id: "feat: git source"
    commit id: "test: git tests"
    checkout main
    merge feature/scan-git id: "merge scan-git"
    branch feature/container-scan
    commit id: "feat: container source"
    commit id: "feat: layer parsing"
    checkout main
    merge feature/container-scan id: "merge container"
    commit id: "release v0.2.0"
```

- `main` — her zaman kararlı ve deploy edilebilir
- `feature/<isim>` — her özellik için ayrı dal
- `fix/<isim>` — hata düzeltmeleri için
- `docs/<isim>` — dokümantasyon güncellemeleri için

### 5.2 Commit Mesajı Formatı: Conventional Commits

```
<tür>[kapsam]: <açıklama>

[gövde]

[alt bilgi]
```

**Türler:**

| Tür | Açıklama |
|-----|----------|
| `feat` | Yeni özellik |
| `fix` | Hata düzeltme |
| `docs` | Dokümantasyon |
| `test` | Test ekleme/düzeltme |
| `refactor` | Yeniden yapılandırma |
| `perf` | Performans iyileştirmesi |
| `ci` | CI/CD değişiklikleri |
| `chore` | Bakım işleri |

**Örnekler:**

```
feat(detector): AWS Secret Access Key dedektörü eklendi
fix(engine): worker pool context iptalinde goroutine sızıntısı düzeltildi
docs(readme): kurulum talimatları güncellendi
test(entropy): Shannon entropi edge case testleri eklendi
perf(matcher): Aho-Corasick otomat derleme süresi %40 iyileştirildi
```

### 5.3 Pull Request Kuralları

- Her PR en az 1 onay (review) gerektirir
- CI pipeline başarılı olmalıdır
- Test kapsamı düşmemelidir
- Linter uyarıları düzeltilmelidir
- PR açıklaması şunları içermelidir:
  - Ne yapıldı ve neden
  - Test planı
  - Breaking change varsa belirtilmeli

### 5.4 Sürüm Numaralama: Semantic Versioning (SemVer)

```
v{MAJOR}.{MINOR}.{PATCH}

MAJOR — Geriye uyumsuz API değişiklikleri
MINOR — Geriye uyumlu yeni özellikler
PATCH — Geriye uyumlu hata düzeltmeleri

Örnekler:
v0.1.0 — İlk MVP sürümü
v0.2.0 — Git entegrasyonu eklendi
v0.3.0 — Doğrulama modülü eklendi
v1.0.0 — Kararlı API, üretime hazır
```

---

## 6. Linter Yapılandırması (.golangci.yml)

```yaml
linters:
  enable:
    - errcheck        # Kontrol edilmeyen hata döndüren fonksiyonlar
    - govet           # Go vet kontrolleri
    - staticcheck     # Gelişmiş statik analiz
    - unused          # Kullanılmayan kod
    - gosimple        # Basitleştirilebilir kod
    - ineffassign     # Etkisiz atamalar
    - typecheck       # Tip kontrolleri
    - gocritic        # Ek stil ve performans kontrolleri
    - gofumpt         # Strict formatlama
    - misspell        # İngilizce yazım kontrolleri
    - prealloc        # Dilim ön-tahsis fırsatları
    - revive          # Ek linting kuralları
    - unconvert       # Gereksiz tip dönüşümleri
    - bodyclose       # HTTP response body kapatma kontrolü
    - noctx           # HTTP isteklerinde context kontrolü

linters-settings:
  gocritic:
    enabled-tags:
      - diagnostic
      - style
      - performance
  revive:
    rules:
      - name: exported
        arguments:
          - "checkPrivateReceivers"

issues:
  exclude-rules:
    - path: _test\.go
      linters:
        - errcheck
        - gocritic
```

---

## 7. Dokümantasyon Standartları

### 7.1 Kod Dokümantasyonu

```go
// Package detector sır tespiti için dedektör arayüzlerini ve
// yerleşik dedektör implementasyonlarını sağlar.
package detector

// AWSAccessKeyID, AWS Access Key ID'lerini tespit eden dedektördür.
// AKIA, ABIA, ACCA ve ASIA ön-ekli anahtarları tanır.
//
// AWS Access Key ID formatı: (AKIA|ABIA|ACCA|ASIA)[0-9A-Z]{16}
//
// Referans: https://docs.aws.amazon.com/IAM/latest/UserGuide/reference_identifiers.html
type AWSAccessKeyID struct{}
```

### 7.2 Proje Dokümantasyonu

| Dosya | İçerik |
|-------|--------|
| `README.md` | Proje tanıtımı, hızlı başlangıç, temel kullanım |
| `docs/01-COMPETITIVE-ANALYSIS.md` | Rakip analizi ve pazar konumlandırma |
| `docs/02-TECHNOLOGY-DECISIONS.md` | Teknoloji kararları ve gerekçeleri |
| `docs/03-ARCHITECTURE.md` | Detaylı mimari tasarım |
| `docs/04-DEVELOPMENT-STANDARDS.md` | Bu belge — geliştirme standartları |
| `docs/05-ROADMAP.md` | Fazlandırılmış geliştirme yol haritası |
| `CONTRIBUTING.md` | Katkıda bulunma rehberi |
| `CHANGELOG.md` | Sürüm değişiklik kayıtları |
| `LICENSE` | MIT Lisansı |

---

## 8. Bağımlılık Yönetimi

### 8.1 Kurallar

- Go modules (`go.mod`) kullanılır
- Bağımlılıklar minimum tutulur — standart kütüphane tercih edilir
- Her bağımlılık ekleme/güncelleme PR açıklamasında gerekçelendirilir
- `govulncheck` ile düzenli güvenlik taraması yapılır
- Doğrudan bağımlılıklar `go.mod`'da açıkça listelenir

### 8.2 Doğrudan Bağımlılık Listesi (Planlanan)

| Bağımlılık | Amaç | Lisans |
|------------|------|--------|
| `github.com/spf13/cobra` | CLI çerçevesi | Apache-2.0 |
| `github.com/spf13/viper` | Yapılandırma yönetimi | MIT |
| `github.com/go-git/go-git/v5` | Git işlemleri | Apache-2.0 |
| `github.com/google/go-containerregistry` | Container imaj | Apache-2.0 |
| `github.com/cloudflare/ahocorasick` | Desen eşleştirme | BSD-3 |
| `github.com/owenrumney/go-sarif` | SARIF çıktı | MIT |
| `github.com/aws/aws-sdk-go-v2` | AWS doğrulama | Apache-2.0 |
| `github.com/stretchr/testify` | Test assertion'ları | MIT |
| `golang.org/x/time` | Rate limiting | BSD-3 |

Tüm bağımlılıklar ticari kullanıma uygun açık kaynak lisanslara sahiptir.

---

## 9. Güvenlik Standartları

### 9.1 Kod Güvenliği

- OWASP Top 10 farkındalığı ile geliştirme
- Kullanıcı girdileri doğrulanır (dosya yolları, URL'ler, regex desenleri)
- Path traversal koruması (`filepath.Clean`, `filepath.Rel`)
- Regex ReDoS koruması (RE2 motoru ile garanti)
- Secrets asla loglanmaz, asla diske yazılmaz (geçici bile olsa)
- `govulncheck` CI pipeline'da zorunlu

### 9.2 Dağıtım Güvenliği

- Release binary'leri checksum ile doğrulanabilir
- GoReleaser ile tekrarlanabilir (reproducible) build
- GitHub Actions'da minimum yetki (principle of least privilege)
- Bağımlılık lisans uyumluluk kontrolü

---

## 10. Performans Profilleme

### 10.1 Profilleme Araçları

```bash
# CPU profili
go test -cpuprofile=cpu.out -bench=BenchmarkScan ./internal/engine/
go tool pprof cpu.out

# Bellek profili
go test -memprofile=mem.out -bench=BenchmarkScan ./internal/engine/
go tool pprof mem.out

# Trace
go test -trace=trace.out -bench=BenchmarkScan ./internal/engine/
go tool trace trace.out
```

### 10.2 Benchmark Testleri

```go
func BenchmarkAhoCorasickMatch(b *testing.B) {
    matcher := NewAhoCorasickMatcher(allKeywords)
    data := loadTestCorpus(b)
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        matcher.Match(data)
    }
}

func BenchmarkShannonEntropy(b *testing.B) {
    data := []byte("AKIAIOSFODNN7EXAMPLE")
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        entropy.Calculate(data)
    }
}
```
