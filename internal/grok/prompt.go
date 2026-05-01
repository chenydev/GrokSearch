package grok

import (
	"strings"
	"time"
)

const searchPrompt = `You are a web search assistant. Answer the user's query directly.

Requirements:
- Prefer authoritative and current sources.
- Keep the answer concise and factual.
- When possible, end with a "Sources:" section containing Markdown links.
- Do not invent citations or URLs.`

func localTimeContext() string {
	now := time.Now()
	return "[Current Time Context]\n- Date: " + now.Format("2006-01-02") + "\n- Time: " + now.Format("15:04:05 MST")
}

func needsTimeContext(query string) bool {
	q := strings.ToLower(query)
	keywords := []string{
		"当前", "现在", "今天", "明天", "昨天", "本周", "上周", "下周",
		"本月", "上月", "下月", "今年", "去年", "明年", "最新", "最近",
		"近期", "实时", "目前",
		"current", "now", "today", "tomorrow", "yesterday", "this week",
		"last week", "next week", "this month", "last month", "next month",
		"this year", "last year", "next year", "latest", "recent", "recently",
		"real-time", "realtime", "up-to-date",
	}
	for _, keyword := range keywords {
		if strings.Contains(q, strings.ToLower(keyword)) {
			return true
		}
	}
	return false
}
