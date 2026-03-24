# ADR-0001: Programlama Dili Seçimi — Go (Golang)

- **Durum:** Kabul Edildi
- **Tarih:** 2026-03-24
- **Karar Verenler:** Proje ekibi

## Bağlam

Leakwatch, çoklu kaynaklarda (Git geçmişi, dosya sistemi, container imajları) sır taraması yapan bir güvenlik aracıdır. Aracın temel gereksinimleri:

- Yüksek performanslı desen eşleştirme (binlerce regex deseni, gigabaytlarca veri)
- Paralel I/O-bound tarama (eşzamanlılık)
- Platformlar arası tek binary dağıtım (Linux, macOS, Windows)
- Git, container imaj ve bulut hizmetleri ile programatik etkileşim
- Güvenlik topluluğu tarafından benimseme potansiyeli

## Karar

**Go (Golang)** birincil geliştirme dili olarak seçilmiştir.

### Gerekçe

1. **Kanıtlanmış alan uygunluğu:** TruffleHog (~17K star) ve Gitleaks (~18K star) aynı problem alanında Go ile yazılmıştır. Mimari referanslar incelenip iyileştirilebilir.

2. **Eşsiz ekosistem birlikteliği:** `go-git` (saf Go git), `go-containerregistry` (OCI/Docker endüstri standardı) ve `cobra`+`viper` (CLI altın standardı) üçlüsü başka hiçbir dilde mevcut değildir.

3. **Eşzamanlılık basitliği:** Goroutine + channel ile fan-out/fan-in desenleri doğal ve hata yapılması zordur.

4. **Dağıtım mükemmelliği:** `GOOS=linux GOARCH=amd64 go build` ile CGO'suz tek statik binary. CI/CD ortamlarında sıfır bağımlılık.

5. **Geliştirme hızı:** Hızlı derleme, basit dil semantiği, geniş geliştirici havuzu.

### Bilinen zayıflık ve azaltma stratejisi

Go'nun RE2 tabanlı `regexp` paketi Rust'ın regex crate'ine kıyasla 2-5x yavaştır. Bu, Aho-Corasick öncelikli hibrit strateji ile azaltılacaktır (bkz. [ADR-0005](ADR-0005-desen-eslestirme.md)). Regex iş yükü %90+ azaltılarak Go'nun regex dezavantajı pratikte ortadan kaldırılır.

## Değerlendirilen Alternatifler

### Rust

- **Artılar:** En iyi regex ve Aho-Corasick performansı, bellek güvenliği
- **Eksiler:** Container imaj kütüphaneleri olgun değil, geliştirme hızı düşük, topluluk katkı bariyeri yüksek, referans mimari yok
- **Karar:** Reddedildi. Gelecekte tarama hot path'i Rust FFI ile hızlandırılabilir.

### Python

- **Artılar:** En geniş geliştirici havuzu, detect-secrets referansı
- **Eksiler:** CPU-bound taramada 10-100x yavaş (GIL), tek binary dağıtım zor, bellek verimsiz
- **Karar:** Reddedildi. Sadece eklenti/kural katmanı için düşünülebilir.

### .NET (C#)

- **Artılar:** İyi performans (AOT), güçlü SARIF desteği
- **Eksiler:** `go-containerregistry` eşdeğeri yok, güvenlik OSS topluluğu zayıf, referans proje yok
- **Karar:** Reddedildi.

### TypeScript

- **Artılar:** Geniş geliştirici havuzu
- **Eksiler:** CPU-bound taramada yavaş, tek binary zor, container/git kütüphaneleri zayıf
- **Karar:** Reddedildi.

## Sonuçlar

### Olumlu

- Kanıtlanmış referans mimariler (TruffleHog, Gitleaks) incelenebilir
- Go ekosisteminin tüm kritik kütüphaneleri mevcut
- Platformlar arası tek binary dağıtım garanti
- Geniş geliştirici havuzu, topluluk katkı potansiyeli
- CI/CD entegrasyonu basit

### Olumsuz

- Regex performansı Rust'a kıyasla düşük (Aho-Corasick ile azaltılacak)
- Generics desteği Go 1.18+ ile geldi, ancak hâlâ Rust/C# kadar olgun değil
- Go'nun hata yönetimi (`if err != nil`) verbose olabilir

## İlişkili Kararlar

- [ADR-0005: Desen Eşleştirme Stratejisi](ADR-0005-desen-eslestirme.md) — Go regex zayıflığının azaltılması
