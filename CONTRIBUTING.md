# Katkıda Bulunma Rehberi

Leakwatch'a katkıda bulunmak istediğiniz için teşekkürler!

## Geliştirme Ortamı

```bash
git clone https://github.com/cemililik/Leakwatch.git
cd Leakwatch
go mod download
go test -race ./...
```

## Gereksinimler

- Go 1.23+
- golangci-lint 1.57+
- Git 2.30+

## Geliştirme Akışı

1. `main` dalından feature branch oluşturun: `git checkout -b feature/my-feature`
2. Değişikliklerinizi yapın
3. Testleri çalıştırın: `go test -race ./...`
4. Lint kontrolü yapın: `golangci-lint run ./...`
5. Pull request oluşturun

## Standartlar

Lütfen aşağıdaki standart belgelerini inceleyin:

- [Geliştirme Standartları](docs/standards/04-DEVELOPMENT-STANDARDS.md)
- [Kod İnceleme Standartları](docs/standards/01-CODE-REVIEW-STANDARDS.md)
- [Dokümantasyon Standartları](docs/standards/00-DOCUMENTATION-STANDARDS.md)

## Commit Mesajları

[Conventional Commits](https://www.conventionalcommits.org/) formatını kullanın:

```
feat(detector): AWS Secret Access Key dedektörü eklendi
fix(engine): worker pool context iptalinde goroutine sızıntısı düzeltildi
test(entropy): Shannon entropi edge case testleri eklendi
```

## Lisans

Katkılarınız [MIT Lisansı](LICENSE) altında lisanslanacaktır.
