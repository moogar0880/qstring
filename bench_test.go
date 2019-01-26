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
	for i := 0; i < b.N; i++ {
		var data QueryStruct
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
		for pb.Next() {
			var data QueryStruct
			if err := Unmarshal(query, &data); err != nil {
				b.Fatal(err)
			}
		}
	})
}
