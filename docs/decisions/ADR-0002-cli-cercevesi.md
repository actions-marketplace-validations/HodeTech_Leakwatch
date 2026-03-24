# ADR-0002: CLI Çerçevesi — Cobra + Viper

- **Durum:** Kabul Edildi
- **Tarih:** 2026-03-24
- **Karar Verenler:** Proje ekibi

## Bağlam

Leakwatch, zengin bir komut yapısına sahip bir CLI aracıdır: `scan git`, `scan fs`, `scan image`, `verify aws` gibi iç içe komutlar gereklidir. Yapılandırma; dosya (.leakwatch.yaml), ortam değişkenleri ve komut satırı flag'lerinden hiyerarşik öncelikle okunmalıdır.

## Karar

CLI çerçevesi olarak **spf13/cobra**, yapılandırma yönetimi için **spf13/viper** seçilmiştir.

### Gerekçe

- Kubernetes, GitHub CLI, Hugo tarafından kullanılan endüstri standardı
- İç içe komut desteği (ağaç yapısı)
- POSIX uyumlu flag yönetimi (`-f`, `--flag`)
- Cobra ↔ Viper doğal entegrasyonu — flag'ler yapılandırma değerlerine bağlanır
- `cobra-cli` ile proje iskeleti oluşturma
- Otomatik yardım metinleri, man page ve markdown çıktı

## Değerlendirilen Alternatifler

### urfave/cli

- **Artılar:** Daha basit API, hızlı başlangıç
- **Eksiler:** İç içe komut desteği daha az esnek, Viper entegrasyonu manuel, daha az ekosistem desteği
- **Karar:** Reddedildi. Leakwatch'ın komut karmaşıklığı için yetersiz.

## Sonuçlar

### Olumlu

- Geliştiricilerin aşina olduğu UX (Kubernetes, gh CLI ile aynı kalıplar)
- `cobra-cli` ile standartlaştırılmış proje yapısı
- Yapılandırma hiyerarşisi (flag > env > config dosya > varsayılan) otomatik

### Olumsuz

- Cobra + Viper birlikte nispeten büyük bir bağımlılık ağacı ekler
- Viper'ın bazı edge case'leri (nested config, type coercion) dikkat gerektirir
