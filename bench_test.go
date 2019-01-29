package qstring

import (
	"net/url"
	"testing"
)

// Straight benchmark literal.
func BenchmarkUnmarshal(b *testing.B) {
	query := url.Values{
		"limit":  []string{"10"},
		"page":   []string{"1"},
		"fields": []string{"a", "b", "c"},
	}
	type QueryStruct struct {
		Fields []string
		Limit  int
		Page   int
	}
	b.ReportAllocs()
	b.ResetTimer()

	var data QueryStruct
	for i := 0; i < b.N; i++ {
		if err := Unmarshal(query, &data); err != nil {
			b.Fatal(err)
		}
	}
}

// Parallel benchmark literal.
func BenchmarkRawPLiteral(b *testing.B) {
	query := url.Values{
		"limit":  []string{"10"},
		"page":   []string{"1"},
		"fields": []string{"a", "b", "c"},
	}
	type QueryStruct struct {
		Fields []string
		Limit  int
		Page   int
	}
	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		var data QueryStruct
		for pb.Next() {
			if err := Unmarshal(query, &data); err != nil {
				b.Fatal(err)
			}
		}
	})
}
