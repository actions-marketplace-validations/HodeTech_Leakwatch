// Package entropy, Shannon entropi hesaplama fonksiyonlarını sağlar.
package entropy

import "math"

// Calculate, verilen byte dizisinin Shannon entropisini hesaplar.
// Sonuç 0.0 (tamamen düzgün) ile ~8.0 (tamamen rastgele) arasındadır.
func Calculate(data []byte) float64 {
	if len(data) == 0 {
		return 0.0
	}

	var freq [256]int
	for _, b := range data {
		freq[b]++
	}

	length := float64(len(data))
	entropy := 0.0
	for _, count := range freq {
		if count == 0 {
			continue
		}
		p := float64(count) / length
		entropy -= p * math.Log2(p)
	}
	return entropy
}
