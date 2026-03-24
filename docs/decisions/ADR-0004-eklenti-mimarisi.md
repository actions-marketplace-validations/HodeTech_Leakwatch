# ADR-0004: Eklenti Mimarisi — Derleme Zamanı Kayıt

- **Durum:** Kabul Edildi
- **Tarih:** 2026-03-24
- **Karar Verenler:** Proje ekibi

## Bağlam

Leakwatch'ın yeni sır dedektörleri, tarama kaynakları ve doğrulayıcıları kolayca eklenebilir olması gerekmektedir. Go'da eklenti mimarisi için iki yaklaşım mevcuttur: çalışma zamanı (runtime) plugin'ler ve derleme zamanı (compile-time) kayıt.

## Karar

**Derleme zamanı kayıt modeli** (compile-time registration via `init()` + blank import) seçilmiştir.

### Gerekçe

- Go'nun `init()` fonksiyonu ve blank import (`import _ "pkg"`) ile standart kayıt deseni
- `database/sql` ve `image` paketlerinde kullanılan kanıtlanmış idiomik Go deseni
- Statik binary felsefesiyle tam uyum
- Kontrolsüz/kötü niyetli eklenti riski yok (güvenlik aracı için kritik)
- Derleme zamanı tip güvenliği

### Mekanizma

1. Her eklenti paketi, belirli bir arayüzü (`Detector`, `Source`, `Verifier`) uygular
2. Paketin `init()` fonksiyonu, kendisini merkezi bir registry'ye kaydeder
3. Ana uygulama, eklenti paketlerini `import _ "pkg"` ile dahil eder
4. Topluluk katkıları Pull Request ile yeni paketler ekleyerek yapılır

## Değerlendirilen Alternatifler

### Çalışma zamanı plugin (Go plugin paketi)

- **Artılar:** Kullanıcılar .so dosyalarını bırakarak eklenti ekleyebilir
- **Eksiler:** Go sürümü, derleyici flag'leri ve C toolchain birebir eşleşmeli — son derece kırılgan. Sadece Linux'ta tam destek. Dağıtım ve bakım karmaşık.
- **Karar:** Reddedildi. Go ekosisteminde fiilen kullanışsızdır.

### YAML tabanlı kural tanımlama (ek olarak)

- **Artılar:** Kod yazmadan basit regex kuralları tanımlama
- **Karar:** Kabul edildi — derleme zamanı modelin tamamlayıcısı olarak. Basit regex desenleri YAML ile, gelişmiş doğrulama mantığı Go arayüzü ile tanımlanır.

## Sonuçlar

### Olumlu

- Güvenli: kontrol edilmemiş kod çalıştırılmaz
- Basit: standart Go import mekanizması, ek araç gerektirmez
- Tip güvenli: derleme zamanında hata yakalanır
- İki katmanlı genişletilebilirlik: YAML (basit) + Go (gelişmiş)

### Olumsuz

- Üçüncü taraf eklentileri fork veya özel derleme gerektirir
- Yeni eklenti eklemek recompile gerektirir (YAML kuralları hariç)

## İlişkili Kararlar

- [ADR-0001: Programlama Dili](ADR-0001-programlama-dili.md) — Go'nun statik derleme felsefesi bu kararı destekler
