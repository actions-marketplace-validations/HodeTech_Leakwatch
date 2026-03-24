# ADR-0005: Desen Eşleştirme Stratejisi — Aho-Corasick Hibrit

- **Durum:** Kabul Edildi
- **Tarih:** 2026-03-24
- **Karar Verenler:** Proje ekibi

## Bağlam

Sır tarama, binlerce farklı desen (regex) ile potansiyel olarak gigabaytlarca veriyi eşleştirmeyi gerektirir. Go'nun RE2 tabanlı `regexp` paketi, Rust'ın regex crate'ine kıyasla 2-5x yavaştır. Her desen için metni tekrar tekrar tarayan naif yaklaşım pratik değildir.

## Karar

**Aho-Corasick öncelikli hibrit strateji** seçilmiştir:

1. **Birincil:** Aho-Corasick algoritması ile sabit keyword ön-filtreleme (tek geçiş, O(n))
2. **İkincil:** Yalnızca Aho-Corasick eşleşmesi olan chunk'larda, sadece eşleşen dedektörlerin regex doğrulaması
3. **Üçüncül:** Shannon entropisi ile ek güven skorlaması

### Gerekçe

- Aho-Corasick, tüm desenleri tek bir geçişte eşleştirir — desen sayısından bağımsız O(n)
- Sır desenlerin %90+'ı sabit ön-eklerle başlar (`AKIA`, `ghp_`, `sk-live-`, `xoxb-`)
- Eşleşme olmayan chunk'lar (%90+ olması beklenir) hiç regex çalıştırmadan atlanır
- Bu yaklaşım, Go'nun regex dezavantajını pratikte ortadan kaldırır
- CPU cache dostu — metin üzerinde tek geçiş

### Kütüphane

`cloudflare/ahocorasick` — Cloudflare üretiminde kanıtlanmış implementasyon.

## Değerlendirilen Alternatifler

### Her desen için ayrı regex (naif)

- **Artılar:** Basit implementasyon
- **Eksiler:** O(n * m) karmaşıklık (n=metin, m=desen sayısı), ölçeklenmez
- **Karar:** Reddedildi.

### Rust FFI ile regex hot path

- **Artılar:** En yüksek ham regex performansı
- **Eksiler:** CGO gerektirir, çapraz derleme karmaşıklaşır, bakım yükü artar
- **Karar:** Ertelendi. Aho-Corasick stratejisi yetersiz kalırsa gelecekte değerlendirilecek.

### Hyperscan (Intel)

- **Artılar:** SIMD-hızlandırılmış çoklu desen eşleştirme
- **Eksiler:** C kütüphanesi (CGO gerekir), Intel'e özgü SIMD, lisans kısıtlamaları
- **Karar:** Reddedildi. Platform bağımlılığı kabul edilemez.

## Sonuçlar

### Olumlu

- Regex iş yükü %90+ azaltılır
- Desen sayısı arttıkça performans sabit kalır (binlerce dedektör eklenebilir)
- CPU cache verimli kullanılır
- Saf Go — CGO gerekmez

### Olumsuz

- Aho-Corasick otomatonunun derlenmesi başlangıçta ek süre gerektirir (ihmal edilebilir)
- Keyword'ü olmayan dedektörler (salt entropi tabanlı) Aho-Corasick ön-filtrelemeden yararlanamaz

## İlişkili Kararlar

- [ADR-0001: Programlama Dili](ADR-0001-programlama-dili.md) — Go regex zayıflığının bu kararı tetiklemesi
