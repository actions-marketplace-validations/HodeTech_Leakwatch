# Architecture Decision Records (ADR)

Bu dizin, Leakwatch projesinin mimari kararlarını [ADR (Architecture Decision Record)](https://adr.github.io/) formatında barındırır.

## ADR Nedir?

ADR, yazılım mimarisinde alınan önemli kararların bağlamını, gerekçesini ve sonuçlarını belgeleyen kısa bir dokümandır. Amacı, gelecekte "bu kararı neden aldık?" sorusuna yanıt vermektir.

## Format

Her ADR aşağıdaki yapıyı izler:

- **Başlık:** `ADR-NNNN: <Karar Başlığı>`
- **Durum:** Önerilen | **Kabul Edildi** | Değiştirildi | Reddedildi | Kullanımdan Kaldırıldı
- **Bağlam:** Kararın alınmasına yol açan durum ve sorun
- **Karar:** Alınan karar ve gerekçesi
- **Değerlendirilen Alternatifler:** İncelenen seçenekler ve reddedilme gerekçeleri
- **Sonuçlar:** Kararın olumlu ve olumsuz etkileri

## Dizin

| ADR | Başlık | Durum | Tarih |
|-----|--------|-------|-------|
| [ADR-0001](ADR-0001-programlama-dili.md) | Programlama Dili: Go | Kabul Edildi | 2026-03-24 |
| [ADR-0002](ADR-0002-cli-cercevesi.md) | CLI Çerçevesi: Cobra + Viper | Kabul Edildi | 2026-03-24 |
| [ADR-0003](ADR-0003-git-kutuphanesi.md) | Git Kütüphanesi: go-git | Kabul Edildi | 2026-03-24 |
| [ADR-0004](ADR-0004-eklenti-mimarisi.md) | Eklenti Mimarisi: Derleme Zamanı | Kabul Edildi | 2026-03-24 |
| [ADR-0005](ADR-0005-desen-eslestirme.md) | Desen Eşleştirme: Aho-Corasick Hibrit | Kabul Edildi | 2026-03-24 |
| [ADR-0006](ADR-0006-container-kutuphanesi.md) | Container Kütüphanesi: go-containerregistry | Kabul Edildi | 2026-03-24 |
| [ADR-0007](ADR-0007-lisans.md) | Lisans: MIT | Kabul Edildi | 2026-03-24 |
| [ADR-0008](ADR-0008-eszamanlilik-modeli.md) | Eşzamanlılık: Worker Pool | Kabul Edildi | 2026-03-24 |
