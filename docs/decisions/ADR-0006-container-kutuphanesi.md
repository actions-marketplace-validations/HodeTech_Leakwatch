# ADR-0006: Container Kütüphanesi — go-containerregistry

- **Durum:** Kabul Edildi
- **Tarih:** 2026-03-24
- **Karar Verenler:** Proje ekibi

## Bağlam

Container imajları, genellikle hassas yapılandırma verileri ve sırlar barındırır. Sırlar, imajın son halinden silinmiş olsa bile önceki katmanlarda var olmaya devam edebilir. Leakwatch, imaj katmanlarını ayrı ayrı inceleyerek "gizlenmiş" sırları ortaya çıkarabilmelidir.

## Karar

**google/go-containerregistry** kütüphanesi seçilmiştir.

### Gerekçe

- Docker daemon gerektirmez — hafif, taşınabilir, CI/CD ortamlarında çalışır
- OCI ve Docker manifest formatlarını tam destekler
- Katman bazlı analiz — her katmanı ayrı ayrı inceleme
- crane, ko, cosign gibi endüstri araçları tarafından kullanılıyor
- Registry kimlik doğrulaması (Docker Hub, GHCR, ECR, GCR)
- Aktif olarak bakımda, Google destekli

## Değerlendirilen Alternatifler

### Docker SDK (docker/docker)

- **Artılar:** Docker API'sine tam erişim
- **Eksiler:** Çalışan Docker daemon gerektirir, ağır bağımlılık
- **Karar:** Reddedildi. Daemon bağımlılığı taşınabilirliği kısıtlar.

### Manuel tar/gzip ayrıştırma

- **Artılar:** Sıfır bağımlılık
- **Eksiler:** Registry auth, manifest parsing, katman yönetimi baştan yazılmalı — büyük iş yükü
- **Karar:** Reddedildi.

## Sonuçlar

### Olumlu

- Daemon'sız çalışma: daha hafif, daha güvenli
- Katman bazlı analiz: silinen sırları önceki katmanlarda tespit
- Çoklu registry desteği: Docker Hub, GHCR, ECR, GCR, özel registry'ler

### Olumsuz

- Büyük imajların katmanlarını indirmek ağ bant genişliği gerektirir
- Bazı özel registry yapılandırmalarında kimlik doğrulama komplikasyonları olabilir
