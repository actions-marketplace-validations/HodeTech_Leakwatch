package detector

import (
	"sort"
	"sync"
)

var (
	mu        sync.RWMutex
	detectors = make(map[string]Detector)
)

// Register, bir dedektörü merkezi kayıt defterine kaydeder.
// Her dedektör paketi, init() fonksiyonunda bu fonksiyonu çağırır.
// Aynı ID ile tekrar kayıt yapılırsa panic oluşur.
func Register(d Detector) {
	mu.Lock()
	defer mu.Unlock()
	if _, exists := detectors[d.ID()]; exists {
		panic("duplicate detector ID: " + d.ID())
	}
	detectors[d.ID()] = d
}

// All, kayıtlı tüm dedektörleri ID'ye göre sıralı şekilde döndürür.
func All() []Detector {
	mu.RLock()
	defer mu.RUnlock()
	result := make([]Detector, 0, len(detectors))
	for _, d := range detectors {
		result = append(result, d)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].ID() < result[j].ID()
	})
	return result
}

// Get, belirtilen ID'ye sahip dedektörü döndürür.
func Get(id string) (Detector, bool) {
	mu.RLock()
	defer mu.RUnlock()
	d, ok := detectors[id]
	return d, ok
}

// Reset, tüm kayıtlı dedektörleri temizler. Sadece testlerde kullanılır.
func Reset() {
	mu.Lock()
	defer mu.Unlock()
	detectors = make(map[string]Detector)
}
