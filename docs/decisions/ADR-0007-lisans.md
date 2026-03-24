# ADR-0007: Lisans — MIT

- **Durum:** Kabul Edildi
- **Tarih:** 2026-03-24
- **Karar Verenler:** Proje ekibi

## Bağlam

Lisans seçimi, projenin benimsenmesini, topluluk katkılarını ve gelecekteki ticari olasılıkları doğrudan etkiler. Sır tarama pazarındaki mevcut durumda:

- TruffleHog **AGPL-3.0** kullanıyor — kurumsal kullanıcıları caydıran güçlü copyleft
- Gitleaks **MIT** kullanıyor — ancak GitHub Action'ı özel repolar için ticari
- GitGuardian tamamen ticari

## Karar

**MIT License** seçilmiştir.

### Gerekçe

1. **Kurumsal benimseme engeli sıfır:** AGPL'den kaçınan birçok kuruluş (bankalar, savunma, büyük teknoloji firmaları) MIT lisanslı araçları tercih eder
2. **Pazar farklılaştırma:** "MIT + doğrulama" kombinasyonu açık kaynak pazarında benzersiz
3. **Open-core model uygunluğu:** Çekirdek MIT olarak kalırken, gelecekte SaaS/Enterprise katmanı eklenebilir
4. **Topluluk katkısı teşviki:** Minimum kısıtlama, maksimum esneklik
5. **Embedding/entegrasyon:** Diğer araçlara gömme veya entegrasyon senaryolarında kısıtlama yok

## Değerlendirilen Alternatifler

### AGPL-3.0

- **Artılar:** Kod değişikliklerinin paylaşılmasını zorunlu kılar, ücretsiz SaaS kullanımını engeller
- **Eksiler:** Birçok kurumsal kuruluş AGPL'yi policy olarak yasaklar; benimseme bariyeri yüksek
- **Karar:** Reddedildi. Leakwatch'ın konumlandırması TruffleHog'un AGPL zayıflığını hedefliyor.

### Apache 2.0

- **Artılar:** Patent koruması içerir, kurumsal dostu
- **Eksiler:** MIT'ye göre daha karmaşık, fark pratikte minimal
- **Karar:** Reddedildi. MIT'nin basitliği ve yaygınlığı tercih edildi.

### BSL (Business Source License)

- **Artılar:** Ticari SaaS kullanımını kısıtlar, sonra açık kaynağa geçer
- **Eksiler:** "Gerçek" açık kaynak olarak kabul görmez (OSI onaylı değil), topluluk güveni zedeler
- **Karar:** Reddedildi.

## Sonuçlar

### Olumlu

- AGPL'den kaçınan kurumsal kullanıcılar için doğal bir seçenek
- Topluluk katkı bariyeri minimum
- Gelecekteki ticari model (SaaS katmanı) ile uyumlu

### Olumsuz

- Rakipler kodu fork edip ticari ürün oluşturabilir
- SaaS olarak sunulması engellenemez (AGPL'de olduğu gibi)
- Bu risk, güçlü bir marka ve topluluk ile azaltılır
