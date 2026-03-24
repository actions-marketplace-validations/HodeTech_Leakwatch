# Leakwatch - Teknoloji Kararları ve Gerekçeleri

> **Belge Versiyonu:** 1.0
> **Tarih:** 2026-03-24
> **Durum:** Taslak

---

## 1. Yönetici Özeti

Bu belge, Leakwatch projesinin teknoloji seçimlerini, bu seçimlerin arkasındaki gerekçeleri ve değerlendirilen alternatifleri detaylı bir şekilde açıklamaktadır. Her karar, performans, ekosistem uygunluğu, geliştirme hızı ve uzun vadeli sürdürülebilirlik kriterleri çerçevesinde alınmıştır.

**Ana Karar:** Go (Golang) birincil geliştirme dili olarak seçilmiştir.

---

## 2. Programlama Dili Seçimi: Go (Golang)

### 2.1 Değerlendirme Kriterleri

| Kriter | Ağırlık | Açıklama |
|--------|---------|----------|
| Regex/Desen Eşleştirme Performansı | %25 | Tarama motorunun darboğazı |
| Eşzamanlılık Modeli | %20 | Paralel I/O-bound tarama |
| Ekosistem (Git, Container, CLI) | %20 | Kritik kütüphanelerin varlığı |
| Çapraz Derleme & Dağıtım | %15 | Tek binary, sıfır bağımlılık |
| Geliştirme Hızı | %10 | İlk sürüme ulaşma süresi |
| Topluluk & İşe Alım | %10 | Katkıda bulunan bulma kolaylığı |

### 2.2 Dil Karşılaştırma Matrisi

| Kriter | Go | Rust | Python | .NET (C#) | TypeScript |
|--------|-----|------|--------|-----------|------------|
| **Regex Performansı** | İyi (RE2) | **En İyi** | Zayıf | İyi | Orta |
| **Aho-Corasick Kalitesi** | Yeterli | **En İyi** | Zayıf | Yeterli | Zayıf |
| **Eşzamanlılık Kolaylığı** | **En İyi** | İyi (karmaşık) | Zayıf (GIL) | İyi | Zayıf |
| **Git Kütüphanesi** | **Mükemmel** (go-git) | Mükemmel (gitoxide) | İyi | İyi | Zayıf |
| **Container İmaj Kütüphanesi** | **En İyi** (go-containerregistry) | Gelişmekte | Yeterli | Zayıf | Zayıf |
| **Çapraz Derleme** | **En İyi** | İyi | Zayıf | Yeterli | Zayıf |
| **Tek Binary** | **Evet** | **Evet** | Hayır | Evet (AOT) | Hayır |
| **SARIF Desteği** | İyi | Yeterli | İyi | **En İyi** | Temel |
| **Geliştirme Hızı** | **Yüksek** | Orta | Yüksek | Orta | Yüksek |
| **Güvenlik Topluluğu** | Geniş | Büyüyen | **En Geniş** | Dar | Geniş |
| **Kanıtlanmış Referanslar** | **TruffleHog, Gitleaks** | ripgrep | detect-secrets | Yok | Yok |

### 2.3 Go Seçim Gerekçesi

1. **Kanıtlanmış Mimari:** TruffleHog ve Gitleaks, Go'nun bu problem alanı için uygunluğunu doğrulamıştır. Çalışan mimariler incelenip iyileştirilebilir.

2. **En İyi Ekosistem Uyumu:**
   - `go-git` — Saf Go, CGO gerektirmez, tam git geçmiş erişimi
   - `go-containerregistry` — OCI/Docker imaj işleme için endüstri standardı
   - `cobra` + `viper` — CLI çerçevesi için altın standart
   - Bu üç kütüphanenin birlikteliği başka hiçbir dilde yoktur

3. **Eşzamanlılık Basitliği:** Goroutine + channel ile fan-out/fan-in desenleri doğal ve hata yapılması zor.

4. **Dağıtım Mükemmelliği:** `GOOS=linux GOARCH=amd64 go build` ile tüm platformlar için tek statik binary. CI/CD entegrasyonu için kritik.

5. **Geliştirme Hızı:** Hızlı derleme, basit dil, büyük geliştirici havuzu.

### 2.4 Go'nun Bilinen Zayıflığı ve Çözüm Stratejisi

**Sorun:** Go'nun RE2 tabanlı `regexp` paketi, Rust'ın `regex` crate'ine kıyasla 2-5x yavaştır.

**Çözüm Stratejisi (Aho-Corasick Öncelikli Yaklaşım):**

Sır desenlerin çoğunluğu sabit ön-eklerle başlar (örn: `AKIA`, `ghp_`, `sk-live-`). Strateji:

1. **Birincil:** Aho-Corasick algoritması ile sabit ön-ek eşleştirme (O(n) — metin boyutuna bağlı, desen sayısından bağımsız)
2. **İkincil:** Yalnızca Aho-Corasick eşleşmesi bulunduğunda regex doğrulaması
3. **Üçüncül:** Entropi analizi ile ek filtreleme

Bu yaklaşım, regex iş yükünü %90+ azaltarak Go'nun regex dezavantajını pratikte ortadan kaldırır.

### 2.5 Neden Rust Değil?

Rust, ham performans açısından en iyi seçim olurdu. Ancak:

- Container imaj kütüphaneleri Go kadar olgun değil
- Geliştirme hızı daha düşük (ownership model öğrenme eğrisi)
- Mevcut referans mimari yok (TruffleHog/Gitleaks Go'da)
- Topluluk katkıları için daha yüksek giriş bariyeri

**Gelecek Olasılık:** Performans kritik hale gelirse, tarama motorunun hot path'i Rust ile yazılıp CGO üzerinden çağrılabilir (hibrit mimari).

### 2.6 Neden .NET Değil?

- Container imaj ayrıştırma için `go-containerregistry` eşdeğeri yok
- Güvenlik OSS topluluğu .NET ekosisteminde çok zayıf
- Referans alınacak benzer proje yok
- Binary boyutları Go/Rust'a göre daha büyük (15-30MB AOT)

---

## 3. Temel Kütüphane Seçimleri

### 3.1 CLI Çerçevesi: Cobra + Viper

| Kütüphane | Versiyon | Amaç |
|-----------|----------|------|
| `github.com/spf13/cobra` | v1.8+ | Komut yapısı, flag yönetimi, yardım metinleri |
| `github.com/spf13/viper` | v1.18+ | Yapılandırma yönetimi (YAML, env vars, flags) |

**Gerekçe:**
- Kubernetes, GitHub CLI, Hugo tarafından kullanılan endüstri standardı
- İç içe komut desteği (`scan git`, `scan fs`, `scan image`, `verify aws`)
- POSIX uyumlu flag yönetimi (`-f`, `--flag`)
- Viper ile sorunsuz entegrasyon — yapılandırma dosyası + ortam değişkeni + flag hiyerarşisi
- `cobra-cli` ile proje iskeletini hızla oluşturma
- Otomatik yardım, man page ve markdown dokümantasyon oluşturma

**Alternatif (Reddedilen):** `urfave/cli` — daha basit projeler için yeterli, ancak iç içe komut desteği ve Viper entegrasyonu yetersiz.

### 3.2 Git İşlemleri: go-git

| Kütüphane | Versiyon | Amaç |
|-----------|----------|------|
| `github.com/go-git/go-git/v5` | v5.12+ | Git repo işlemleri, geçmiş analizi |

**Gerekçe:**
- Saf Go implementasyonu — CGO gerektirmez, çapraz derleme sorunsuz
- Harici `git` binary bağımlılığı yok
- Git nesneleri üzerinde tam programatik kontrol
- TruffleHog tarafından kullanılıyor — kanıtlanmış
- `LogOptions` ile optimize edilmiş tarama (since, depth)
- Pluggable storage ile bellek içi test desteği

**Alternatif (Reddedilen):** `git2go` — C bağımlılığı (libgit2), CGO karmaşıklığı, çapraz derleme zorluğu.

### 3.3 Container İmaj İşlemleri: go-containerregistry

| Kütüphane | Versiyon | Amaç |
|-----------|----------|------|
| `github.com/google/go-containerregistry` | v0.20+ | OCI/Docker imaj katmanlarının analizi |

**Gerekçe:**
- Docker daemon gerektirmez — hafif, taşınabilir
- OCI ve Docker manifest formatlarını destekler
- Katman bazlı analiz — silinen dosyaları önceki katmanlarda tespit
- crane, ko, cosign tarafından kullanılıyor
- Registry kimlik doğrulaması desteği (Docker Hub, GHCR, ECR, GCR)

### 3.4 Desen Eşleştirme: Aho-Corasick

| Kütüphane | Versiyon | Amaç |
|-----------|----------|------|
| `github.com/cloudflare/ahocorasick` | latest | Çoklu desen eşleştirme |

**Gerekçe:**
- O(n) zaman karmaşıklığı — metin boyutuna bağlı, desen sayısından bağımsız
- Binlerce desen eklendiğinde bile performans sabit
- CPU cache dostu — tek geçişli tarama
- Cloudflare'in üretimde kullandığı kanıtlanmış implementasyon

**Alternatif:** `github.com/petar-dambovaliev/aho-corasick` — daha yeni, daha Go-idiomatic API.

### 3.5 Çıktı Formatları

| Kütüphane | Amaç |
|-----------|------|
| `github.com/owenrumney/go-sarif` | SARIF çıktı formatı |
| `encoding/json` (stdlib) | JSON çıktı |
| `encoding/csv` (stdlib) | CSV çıktı |

### 3.6 Test Altyapısı

| Kütüphane | Amaç |
|-----------|------|
| `testing` (stdlib) | Birim testler |
| `testing/fstest` (stdlib) | Bellek içi dosya sistemi testleri |
| `github.com/stretchr/testify` | Assertion ve mock kütüphanesi |

### 3.7 AWS/Cloud SDK'ları (Doğrulama İçin)

| Kütüphane | Amaç |
|-----------|------|
| `github.com/aws/aws-sdk-go-v2` | AWS STS GetCallerIdentity (anahtar doğrulama) |
| `net/http` (stdlib) | GitHub, Slack vb. API doğrulama |

---

## 4. Derleme ve Dağıtım Araçları

### 4.1 GoReleaser

| Araç | Amaç |
|------|------|
| `goreleaser` | Çapraz derleme, arşivleme ve GitHub Release oluşturma |

**Gerekçe:**
- Tek komutla Linux/macOS/Windows (amd64, arm64) için derleme
- Otomatik GitHub Release varlık yükleme
- Homebrew formula ve Scoop manifest oluşturma
- Docker imaj oluşturma ve yayınlama
- Changelog otomatik oluşturma

### 4.2 GitHub Actions

| Workflow | Amaç |
|----------|------|
| `ci.yml` | Her push/PR için test, lint, build |
| `release.yml` | Tag push'ta GoReleaser ile sürüm yayınlama |

### 4.3 Kod Kalitesi Araçları

| Araç | Amaç |
|------|------|
| `golangci-lint` | Statik analiz ve linting (50+ linter) |
| `gofumpt` | Strict Go kod formatlama |
| `govulncheck` | Bilinen güvenlik açığı taraması |

---

## 5. Minimum Go Versiyonu

**Go 1.22+** (tercihen en güncel kararlı sürüm)

**Gerekçe:**
- `io/fs` paketi (Go 1.16+) — dosya sistemi soyutlaması
- Generics desteği (Go 1.18+) — tip güvenli koleksiyonlar
- `log/slog` (Go 1.21+) — yapılandırılmış loglama
- Gelişmiş GC performansı (Go 1.22+)
- `range over func` (Go 1.23+) — iterator desteği

---

## 6. Proje Yapısı

```
leakwatch/
├── cmd/                        # CLI komutları (Cobra)
│   ├── root.go                 # Ana komut
│   ├── scan.go                 # scan üst komutu
│   ├── scan_git.go             # scan git alt komutu
│   ├── scan_fs.go              # scan fs alt komutu
│   ├── scan_image.go           # scan image alt komutu
│   └── verify.go               # verify komutu
├── internal/                   # Dahili paketler (dışa kapalı)
│   ├── engine/                 # Tarama motoru çekirdeği
│   │   ├── engine.go           # İşçi havuzu ve orkestrasyon
│   │   ├── worker.go           # İşçi goroutine'leri
│   │   └── pipeline.go         # Tarama pipeline'ı
│   ├── detector/               # Sır dedektörleri
│   │   ├── registry.go         # Dedektör kayıt defteri
│   │   ├── detector.go         # Detector arayüzü
│   │   ├── aws.go              # AWS dedektörleri
│   │   ├── github.go           # GitHub dedektörleri
│   │   ├── generic.go          # Genel dedektörler
│   │   └── custom.go           # YAML tabanlı özel kurallar
│   ├── source/                 # Tarama kaynakları
│   │   ├── source.go           # Source arayüzü
│   │   ├── git.go              # Git kaynağı
│   │   ├── filesystem.go       # Dosya sistemi kaynağı
│   │   └── container.go        # Container imaj kaynağı
│   ├── verifier/               # Sır doğrulama modülleri
│   │   ├── verifier.go         # Verifier arayüzü
│   │   ├── aws.go              # AWS STS doğrulama
│   │   └── github.go           # GitHub API doğrulama
│   ├── entropy/                # Entropi hesaplama
│   │   └── shannon.go          # Shannon entropi implementasyonu
│   ├── matcher/                # Desen eşleştirme motoru
│   │   ├── ahocorasick.go      # Aho-Corasick implementasyonu
│   │   └── regex.go            # Regex doğrulama
│   ├── output/                 # Çıktı formatlayıcıları
│   │   ├── formatter.go        # Formatter arayüzü
│   │   ├── json.go             # JSON çıktı
│   │   ├── sarif.go            # SARIF çıktı
│   │   └── csv.go              # CSV çıktı
│   ├── config/                 # Yapılandırma yönetimi
│   │   └── config.go           # Viper tabanlı config
│   └── filter/                 # Filtreleme (.leakwatchignore vb.)
│       └── filter.go           # Dosya/yol filtreleme
├── pkg/                        # Dışa açık paketler (kütüphane kullanımı)
│   └── finding/                # Finding veri yapısı
│       └── finding.go          # Bulgu modeli
├── rules/                      # Yerleşik kural tanımları
│   ├── aws.yaml                # AWS sır desenleri
│   ├── github.yaml             # GitHub sır desenleri
│   ├── gcp.yaml                # GCP sır desenleri
│   ├── generic.yaml            # Genel sır desenleri
│   └── ...
├── docs/                       # Proje dokümantasyonu
├── .github/                    # GitHub Actions workflow'ları
│   └── workflows/
│       ├── ci.yml
│       └── release.yml
├── .goreleaser.yml             # GoReleaser yapılandırması
├── .golangci.yml               # Linter yapılandırması
├── .pre-commit-hooks.yaml      # Pre-commit hook tanımı
├── go.mod                      # Go modül tanımı
├── go.sum                      # Bağımlılık checksum'ları
├── main.go                     # Giriş noktası
├── LICENSE                     # MIT Lisansı
└── README.md                   # Proje açıklaması
```

---

## 7. Performans Hedefleri

| Metrik | Hedef | Referans |
|--------|-------|---------|
| Orta boy repo tarama (10K commit) | < 30 saniye | Gitleaks ~60s, TruffleHog ~120s |
| Dosya sistemi tarama (10K dosya) | < 10 saniye | — |
| Container imaj tarama (500MB) | < 60 saniye | — |
| Bellek kullanımı (orta repo) | < 512MB | TruffleHog 1GB+ olabiliyor |
| Binary boyutu | < 30MB | — |
| Başlangıç süresi | < 100ms | — |

---

## 8. Lisans Kararı: MIT

**Gerekçe:**
- Kurumsal benimseme için sıfır engel (AGPL'den kaçınan kullanıcıları hedefleme)
- Gitleaks ile aynı lisans modeli — kanıtlanmış yaklaşım
- Gelecekte ticari katman (SaaS/Enterprise) eklenebilir (open-core model)
- Topluluk katkılarını teşvik eder
- Embedding/entegrasyon senaryolarında kısıtlama yok
