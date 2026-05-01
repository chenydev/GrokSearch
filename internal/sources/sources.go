package sources

import (
	"encoding/json"
	"regexp"
	"strings"

	"github.com/GuDaStudio/GrokSearch/internal/firecrawl"
	"github.com/GuDaStudio/GrokSearch/internal/tavily"
)

type Source struct {
	Title       string `json:"title,omitempty"`
	URL         string `json:"url"`
	Description string `json:"description,omitempty"`
	Provider    string `json:"provider,omitempty"`
}

var (
	urlPattern     = regexp.MustCompile(`https?://[^\s<>"'\x60，。、；：！？》）】)]+`)
	mdLinkPattern  = regexp.MustCompile(`\[([^\]]+)\]\((https?://[^)]+)\)`)
	headingPattern = regexp.MustCompile(`(?im)^\s*(?:#{1,6}\s*)?(?:\*\*|__)?\s*(sources?|references?|citations?|信源|参考资料|参考|引用|来源列表|来源)\s*(?:\*\*|__)?\s*[:：]?\s*$`)
)

func SplitAnswerAndSources(text string) (string, []Source) {
	raw := strings.TrimSpace(text)
	if raw == "" {
		return "", nil
	}
	if answer, srcs, ok := splitHeading(raw); ok {
		return answer, srcs
	}
	if answer, srcs, ok := splitTailLinks(raw); ok {
		return answer, srcs
	}
	return raw, nil
}

func Merge(lists ...[]Source) []Source {
	seen := map[string]bool{}
	var out []Source
	for _, list := range lists {
		for _, src := range list {
			src.URL = strings.TrimSpace(src.URL)
			if src.URL == "" || seen[src.URL] {
				continue
			}
			seen[src.URL] = true
			out = append(out, src)
		}
	}
	return out
}

func FromTavily(results []tavily.SearchResult) []Source {
	out := make([]Source, 0, len(results))
	for _, r := range results {
		if r.URL == "" {
			continue
		}
		out = append(out, Source{Title: r.Title, URL: r.URL, Description: r.Content, Provider: "tavily"})
	}
	return out
}

func FromFirecrawl(results []firecrawl.SearchResult) []Source {
	out := make([]Source, 0, len(results))
	for _, r := range results {
		if r.URL == "" {
			continue
		}
		out = append(out, Source{Title: r.Title, URL: r.URL, Description: r.Description, Provider: "firecrawl"})
	}
	return out
}

func Extract(text string) []Source {
	seen := map[string]bool{}
	var out []Source
	for _, match := range mdLinkPattern.FindAllStringSubmatch(text, -1) {
		title := strings.TrimSpace(match[1])
		url := strings.TrimSpace(match[2])
		if url == "" || seen[url] {
			continue
		}
		seen[url] = true
		out = append(out, Source{Title: title, URL: url})
	}
	for _, raw := range urlPattern.FindAllString(text, -1) {
		url := strings.TrimRight(raw, ".,;:!?")
		if url == "" || seen[url] {
			continue
		}
		seen[url] = true
		out = append(out, Source{URL: url})
	}
	return out
}

func WriteJSON(srcs []Source) ([]byte, error) {
	return json.MarshalIndent(srcs, "", "  ")
}

func splitHeading(text string) (string, []Source, bool) {
	matches := headingPattern.FindAllStringIndex(text, -1)
	for i := len(matches) - 1; i >= 0; i-- {
		start := matches[i][0]
		srcs := Extract(text[start:])
		if len(srcs) == 0 {
			continue
		}
		return strings.TrimSpace(text[:start]), srcs, true
	}
	return "", nil, false
}

func splitTailLinks(text string) (string, []Source, bool) {
	lines := strings.Split(text, "\n")
	idx := len(lines) - 1
	for idx >= 0 && strings.TrimSpace(lines[idx]) == "" {
		idx--
	}
	if idx < 0 {
		return "", nil, false
	}
	end := idx
	count := 0
	for idx >= 0 {
		line := strings.TrimSpace(lines[idx])
		if line == "" {
			idx--
			continue
		}
		if !isLinkOnlyLine(line) {
			break
		}
		count++
		idx--
	}
	if count < 2 {
		return "", nil, false
	}
	block := strings.Join(lines[idx+1:end+1], "\n")
	srcs := Extract(block)
	if len(srcs) == 0 {
		return "", nil, false
	}
	return strings.TrimSpace(strings.Join(lines[:idx+1], "\n")), srcs, true
}

func isLinkOnlyLine(line string) bool {
	line = strings.TrimSpace(line)
	line = regexp.MustCompile(`^\s*(?:[-*]|\d+\.)\s*`).ReplaceAllString(line, "")
	return strings.HasPrefix(line, "http://") || strings.HasPrefix(line, "https://") || mdLinkPattern.MatchString(line)
}
