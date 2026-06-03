package embeddings

import (
	"encoding/binary"
	"hash/fnv"
	"math"
	"strings"
)

// Dim is the fixed embedding dimension.
const Dim = 256

// Compute creates a feature-hashing embedding for text.
func Compute(text string) []float32 {
	vec := make([]float32, Dim)
	words := strings.Fields(strings.ToLower(text))
	for _, w := range words {
		if len(w) < 2 {
			continue
		}
		h := fnv.New32a()
		h.Write([]byte(w))
		idx := h.Sum32() % Dim
		vec[idx] += 1.0
	}
	var norm float64
	for _, v := range vec {
		norm += float64(v) * float64(v)
	}
	norm = math.Sqrt(norm)
	if norm > 0 {
		for i := range vec {
			vec[i] = float32(float64(vec[i]) / norm)
		}
	}
	return vec
}

// ToBytes serializes a float32 vector to bytes.
func ToBytes(v []float32) []byte {
	b := make([]byte, len(v)*4)
	for i, f := range v {
		binary.LittleEndian.PutUint32(b[i*4:], math.Float32bits(f))
	}
	return b
}

// FromBytes deserializes bytes to a float32 vector.
func FromBytes(b []byte) []float32 {
	if len(b)%4 != 0 {
		return nil
	}
	n := len(b) / 4
	v := make([]float32, n)
	for i := 0; i < n; i++ {
		bits := binary.LittleEndian.Uint32(b[i*4:])
		v[i] = math.Float32frombits(bits)
	}
	return v
}

// Cosine similarity between two vectors.
func Cosine(a, b []float32) float64 {
	if len(a) != len(b) {
		return 0
	}
	var dot, normA, normB float64
	for i := range a {
		va := float64(a[i])
		vb := float64(b[i])
		dot += va * vb
		normA += va * va
		normB += vb * vb
	}
	if normA == 0 || normB == 0 {
		return 0
	}
	return dot / (math.Sqrt(normA) * math.Sqrt(normB))
}
