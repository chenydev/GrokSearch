package sources

import "testing"

func TestSplitHeadingSources(t *testing.T) {
	answer, srcs := SplitAnswerAndSources("Answer\n\nSources:\n- [Doc](https://example.com/doc)\n- https://example.com/raw")
	if answer != "Answer" {
		t.Fatalf("answer = %q", answer)
	}
	if len(srcs) != 2 {
		t.Fatalf("sources = %d", len(srcs))
	}
	if srcs[0].Title != "Doc" || srcs[0].URL != "https://example.com/doc" {
		t.Fatalf("unexpected first source: %#v", srcs[0])
	}
}

func TestMergeDeduplicates(t *testing.T) {
	srcs := Merge(
		[]Source{{URL: "https://example.com/a"}, {URL: "https://example.com/b"}},
		[]Source{{URL: "https://example.com/a"}, {URL: "https://example.com/c"}},
	)
	if len(srcs) != 3 {
		t.Fatalf("sources = %d", len(srcs))
	}
}
