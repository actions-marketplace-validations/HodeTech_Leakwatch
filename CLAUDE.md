# CLAUDE.md — Leakwatch Geliştirme Rehberi

Bu dosya, Claude Code'un Leakwatch projesi üzerinde çalışırken referans alması gereken standartları ve bağlamı tanımlar.

## Proje Tanımı

Leakwatch, kod tabanlarında, Git geçmişlerinde ve container imajlarında sızan sırları (API anahtarları, parolalar, sertifikalar) tespit eden, doğrulayan ve raporlayan yüksek performanslı, açık kaynak (MIT) bir güvenlik aracıdır.

**Dil:** Go (1.22+)
**Lisans:** MIT
**Repo:** https://github.com/cemililik/Leakwatch

## Proje Yapısı

```
leakwatch/
├── cmd/                    # CLI komutları (Cobra) — ince katman, iş mantığı yok
├── internal/               # Dahili paketler — tüm iş mantığı burada
│   ├── engine/             # Tarama motoru (worker pool, pipeline)
│   ├── detector/           # Sır dedektörleri (Detector arayüzü)
│   ├── source/             # Tarama kaynakları (Source arayüzü)
│   ├── verifier/           # Sır doğrulama (Verifier arayüzü)
│   ├── entropy/            # Shannon entropi hesaplama
│   ├── matcher/            # Aho-Corasick + regex motoru
│   ├── output/             # Çıktı formatlayıcıları (Formatter arayüzü)
│   ├── config/             # Viper tabanlı yapılandırma
│   └── filter/             # .leakwatchignore, satır içi yoksayma
├── pkg/                    # Dışa açık paketler (finding modeli)
├── rules/                  # YAML sır kural tanımları
├── docs/                   # Dokümantasyon
│   ├── architecture/       # Mimari ve teknik tasarım belgeleri
│   ├── standards/          # Geliştirme ve dokümantasyon standartları
│   ├── decisions/          # ADR (Architecture Decision Records)
│   └── 05-ROADMAP.md       # Yol haritası
└── main.go                 # Giriş noktası
```

## Temel Mimari Kararlar

Mimari kararlar `docs/decisions/` altında ADR formatında belgelenmiştir. Geliştirme sırasında bu kararlara uyulmalıdır:

| ADR | Karar | Özet |
|-----|-------|------|
| [ADR-0001](docs/decisions/ADR-0001-programlama-dili.md) | Go | Kanıtlanmış ekosistem, eşzamanlılık, tek binary |
| [ADR-0002](docs/decisions/ADR-0002-cli-cercevesi.md) | Cobra + Viper | İç içe komutlar, hiyerarşik yapılandırma |
| [ADR-0003](docs/decisions/ADR-0003-git-kutuphanesi.md) | go-git | Saf Go, CGO yok, harici bağımlılık yok |
| [ADR-0004](docs/decisions/ADR-0004-eklenti-mimarisi.md) | Derleme zamanı kayıt | init() + blank import, tip güvenli |
| [ADR-0005](docs/decisions/ADR-0005-desen-eslestirme.md) | Aho-Corasick hibrit | AC ön-filtre → regex doğrulama → entropi |
| [ADR-0006](docs/decisions/ADR-0006-container-kutuphanesi.md) | go-containerregistry | Daemon'suz, katman bazlı analiz |
| [ADR-0007](docs/decisions/ADR-0007-lisans.md) | MIT | Kurumsal benimseme, open-core uyum |
| [ADR-0008](docs/decisions/ADR-0008-eszamanlilik-modeli.md) | Worker Pool | Sabit işçi sayısı, channel tabanlı |

## Kodlama Standartları

Tam standartlar: [docs/standards/04-DEVELOPMENT-STANDARDS.md](docs/standards/04-DEVELOPMENT-STANDARDS.md)

### Kritik Kurallar

- **Dil:** Go 1.22+, `CGO_ENABLED=0`
- **Stil:** Effective Go + Uber Go Style Guide
- **Linting:** `golangci-lint` zorunlu, tüm CI'da çalışır
- **Formatlama:** `gofumpt` (strict gofmt)
- **Test kapsamı:** minimum %80, dedektörler %95
- **Hata yönetimi:** Her hatayı `fmt.Errorf("bağlam: %w", err)` ile sarmalayarak döndür
- **Loglama:** `log/slog` yapılandırılmış loglama — fmt.Println/log.Printf KULLANMA
- **Sır güvenliği:** Bulunan sırları ASLA logla, diske yaz veya önbelleğe al

### Adlandırma

| Öğe | Kural | Örnek |
|-----|-------|-------|
| Paket | Kısa, küçük harf | `detector`, `engine` |
| Dışa açık | PascalCase | `ScanRepository()` |
| Dahili | camelCase | `parseConfig()` |
| Interface | PascalCase, "-er" soneki | `Detector`, `Verifier` |
| Dosya | snake_case | `aws_access_key.go` |
| Test | `_test.go` soneki | `engine_test.go` |

### Paket Kuralları

- `cmd/` → Sadece CLI wiring, iş mantığı yok
- `internal/` → Tüm iş mantığı, dışa kapalı
- `pkg/` → Dışa açık tipler (Finding modeli)
- Standart kütüphane tercih edilir, gereksiz bağımlılık ekleme

### Test Yazımı

- **Table-driven testler** tercih et
- `testing/fstest.MapFS` ile bellek içi dosya sistemi testi
- Test adlandırma: `Test<Fonksiyon>_<Senaryo>_<BeklenenSonuç>`
- Mock'lar: arayüzlere karşı test et, mock'lar doğal gelir
- Race detector: `go test -race ./...`

## Commit Standartları

**Format:** Conventional Commits

```
<tür>(<kapsam>): <açıklama>

Türler: feat, fix, docs, test, refactor, perf, ci, chore
```

**Örnekler:**
```
feat(detector): AWS Secret Access Key dedektörü eklendi
fix(engine): worker pool context iptalinde goroutine sızıntısı düzeltildi
test(entropy): Shannon entropi edge case testleri eklendi
```

## Temel Bağımlılıklar

| Paket | Amaç |
|-------|------|
| `spf13/cobra` | CLI çerçevesi |
| `spf13/viper` | Yapılandırma yönetimi |
| `go-git/go-git/v5` | Git işlemleri |
| `google/go-containerregistry` | Container imaj analizi |
| `cloudflare/ahocorasick` | Çoklu desen eşleştirme |
| `owenrumney/go-sarif` | SARIF çıktı |
| `aws/aws-sdk-go-v2` | AWS doğrulama |
| `stretchr/testify` | Test assertion'ları |

## Dokümantasyon Standartları

Tam standartlar: [docs/standards/00-DOCUMENTATION-STANDARDS.md](docs/standards/00-DOCUMENTATION-STANDARDS.md)

- Tüm diyagramlar **Mermaid** formatında (ASCII art KULLANMA)
- Kod blokları dil etiketi içermeli: ` ```go `, ` ```yaml `
- Dahili bağlantılar göreceli yol kullanır
- Mimari kararlar `docs/decisions/ADR-NNNN-*.md` formatında belgelenir

## Yapılmaması Gerekenler

- Sır içeriğini loglama, konsola yazma veya test fixture'larına koyma
- CGO gerektiren kütüphane ekleme (çapraz derleme kırılır)
- `cmd/` altına iş mantığı koyma
- ASCII art diyagram oluşturma (Mermaid kullan)
- Mevcut ADR'lara aykırı mimari kararlar alma (önce ADR güncelle)
- `go.sum` veya `vendor/` dizinini manuel düzenleme
- `--no-verify` ile git hook'larını atlama
