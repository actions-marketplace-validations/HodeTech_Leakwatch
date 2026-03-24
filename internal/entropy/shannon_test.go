package entropy

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCalculate_EmptyInput_ReturnsZero(t *testing.T) {
	assert.Equal(t, 0.0, Calculate([]byte{}))
}

func TestCalculate_NilInput_ReturnsZero(t *testing.T) {
	assert.Equal(t, 0.0, Calculate(nil))
}

func TestCalculate_SingleByte_ReturnsZero(t *testing.T) {
	assert.Equal(t, 0.0, Calculate([]byte{0x42}))
}

func TestCalculate_RepeatedCharacter_ReturnsZero(t *testing.T) {
	assert.Equal(t, 0.0, Calculate([]byte("aaaaaaaaaa")))
}

func TestCalculate_All256UniqueBytes_ReturnsMaxEntropy(t *testing.T) {
	data := make([]byte, 256)
	for i := 0; i < 256; i++ {
		data[i] = byte(i)
	}
	h := Calculate(data)
	assert.InDelta(t, 8.0, h, 0.001, "256 unique bytes should yield ~8.0 bits entropy")
}

func TestCalculate_TwoCharactersEqualDistribution_ReturnsOne(t *testing.T) {
	h := Calculate([]byte("ababababab"))
	assert.GreaterOrEqual(t, h, 0.9, "entropi alt sınırın altında")
	assert.LessOrEqual(t, h, 1.1, "entropi üst sınırın üstünde")
}

func TestCalculate_LowEntropyWord_ReturnsBetween2And4(t *testing.T) {
	h := Calculate([]byte("password"))
	assert.GreaterOrEqual(t, h, 2.5, "entropi alt sınırın altında")
	assert.LessOrEqual(t, h, 3.5, "entropi üst sınırın üstünde")
}

func TestCalculate_AWSAccessKeyFormat_ReturnsMediumEntropy(t *testing.T) {
	h := Calculate([]byte("AKIAIOSFODNN7EXAMPLE"))
	assert.GreaterOrEqual(t, h, 3.5, "entropi alt sınırın altında")
	assert.LessOrEqual(t, h, 4.5, "entropi üst sınırın üstünde")
}

func TestCalculate_HighEntropyString_ReturnsAbove4(t *testing.T) {
	h := Calculate([]byte("aB3$kL9@mN2!pQ7&rT4^"))
	assert.GreaterOrEqual(t, h, 4.0, "entropi alt sınırın altında")
	assert.LessOrEqual(t, h, 5.0, "entropi üst sınırın üstünde")
}

func BenchmarkCalculate(b *testing.B) {
	// Simulate a typical secret-length input
	data := []byte("AKIAIOSFODNN7EXAMPLE+wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Calculate(data)
	}
}
