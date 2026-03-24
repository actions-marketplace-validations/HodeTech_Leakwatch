# ADR-0003: Git Kütüphanesi — go-git

- **Durum:** Kabul Edildi
- **Tarih:** 2026-03-24
- **Karar Verenler:** Proje ekibi

## Bağlam

Leakwatch, Git depolarının tüm commit geçmişini analiz edebilmelidir. Bu, depoyu açma/klonlama, commit'lerde gezinme, her commit'in dosya ağacını okuma ve blob içeriklerine erişme gerektirir. İki ana yaklaşım mevcuttur: harici `git` binary çağrısı (os/exec) veya Go-native bir kütüphane.

## Karar

**go-git/go-git/v5** (saf Go implementasyonu) seçilmiştir.

### Gerekçe

- Saf Go — CGO gerektirmez, çapraz derleme sorunsuz
- Harici `git` binary bağımlılığı yok — hedef sistemde git kurulu olması gerekmez
- Git nesneleri üzerinde tam programatik kontrol
- TruffleHog tarafından üretimde kullanılıyor — kanıtlanmış
- `LogOptions` ile optimize tarama (`Since`, `Order`, depth sınırlama)
- Pluggable storage ile bellek içi test desteği

## Değerlendirilen Alternatifler

### git2go (libgit2 bindings)

- **Artılar:** libgit2'nin olgun ve kapsamlı API'si
- **Eksiler:** CGO gerektirir, çapraz derleme karmaşıklaşır, C derleyici zinciri eşleşmesi gerekir
- **Karar:** Reddedildi. Go'nun statik binary felsefesine aykırı.

### os/exec ile git komutu

- **Artılar:** Her git özelliğine erişim, basit implementasyon
- **Eksiler:** Harici bağımlılık, metin ayrıştırma (parsing) gerektirir, performans overhead'i, güvenlik riski (command injection)
- **Karar:** Reddedildi. Üretim kalitesinde bir araç için güvenilir değil.

## Sonuçlar

### Olumlu

- Sıfır harici bağımlılık — tek binary, her yerde çalışır
- Git iç yapılarına doğrudan erişim (tree, blob, commit nesneleri)
- Test edilebilirlik: bellek içi repo oluşturma ile birim test

### Olumsuz

- go-git, yerel git'in tüm özelliklerini desteklemez (örn: shallow clone sınırlamaları)
- Çok büyük monorepolarda bellek tüketimi yerel git'e kıyasla yüksek olabilir
- Bazı edge case'lerde (submodule, sparse checkout) davranış farklılıkları
