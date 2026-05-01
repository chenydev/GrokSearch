package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/GuDaStudio/GrokSearch/internal/config"
	"github.com/GuDaStudio/GrokSearch/internal/firecrawl"
	"github.com/GuDaStudio/GrokSearch/internal/grok"
	"github.com/GuDaStudio/GrokSearch/internal/sources"
	"github.com/GuDaStudio/GrokSearch/internal/tavily"
	"github.com/spf13/cobra"
)

type app struct {
	configPath string
	format     string
	timeout    time.Duration
}

type searchOutput struct {
	Content      string           `json:"content"`
	Sources      []sources.Source `json:"sources,omitempty"`
	SourcesCount int              `json:"sources_count"`
	Model        string           `json:"model,omitempty"`
}

type fetchOutput struct {
	URL      string `json:"url"`
	Provider string `json:"provider"`
	Content  string `json:"content"`
}

func NewRootCommand() *cobra.Command {
	a := &app{format: "text", timeout: 120 * time.Second}
	cmd := &cobra.Command{
		Use:           "grok-search",
		Short:         "Grok/Tavily/Firecrawl web search CLI",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	cmd.PersistentFlags().StringVar(&a.configPath, "config", "", "config file path")
	cmd.PersistentFlags().StringVar(&a.format, "format", "text", "output format: text or json")
	cmd.PersistentFlags().DurationVar(&a.timeout, "timeout", 120*time.Second, "request timeout")

	cmd.AddCommand(a.searchCommand())
	cmd.AddCommand(a.fetchCommand())
	cmd.AddCommand(a.mapCommand())
	cmd.AddCommand(a.configCommand())
	cmd.AddCommand(a.modelsCommand())
	cmd.AddCommand(a.lastCommand())
	return cmd
}

func (a *app) searchCommand() *cobra.Command {
	var platform, model, sourcesPath string
	var extraSources int
	var showSources bool

	cmd := &cobra.Command{
		Use:   "search QUERY",
		Short: "Search the web through a Grok-compatible API",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			loaded, err := config.Load(a.configPath)
			if err != nil {
				return err
			}
			cfg := loaded.Config
			if model != "" {
				cfg.GrokModel = model
			}
			query := strings.Join(args, " ")
			ctx, cancel := context.WithTimeout(cmd.Context(), a.timeout)
			defer cancel()

			answer, err := grok.New(cfg, a.timeout).Search(ctx, query, platform, cfg.GrokModel)
			if err != nil {
				return err
			}
			content, grokSources := sources.SplitAnswerAndSources(answer)
			allSources := append([]sources.Source{}, grokSources...)

			if extraSources > 0 {
				extra, err := collectExtraSources(ctx, cfg, query, extraSources, a.timeout)
				if err == nil {
					allSources = sources.Merge(allSources, extra)
				}
			}

			out := searchOutput{
				Content:      content,
				Sources:      allSources,
				SourcesCount: len(allSources),
				Model:        cfg.GrokModel,
			}
			_ = writeLast(out)

			if sourcesPath != "" {
				if err := writeSourcesFile(sourcesPath, allSources); err != nil {
					return err
				}
			}
			if a.format == "json" {
				return printJSON(out)
			}
			fmt.Println(content)
			if showSources {
				printSources(allSources)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&platform, "platform", "", "platform focus, e.g. GitHub or Reddit")
	cmd.Flags().StringVar(&model, "model", "", "model for this request")
	cmd.Flags().IntVar(&extraSources, "extra-sources", 0, "number of extra Tavily/Firecrawl sources to collect")
	cmd.Flags().StringVar(&sourcesPath, "sources", "", "write sources JSON to path")
	cmd.Flags().BoolVar(&showSources, "show-sources", false, "print sources after text output")
	return cmd
}

func (a *app) fetchCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fetch URL",
		Short: "Fetch a web page as Markdown through Tavily, with Firecrawl fallback",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			loaded, err := config.Load(a.configPath)
			if err != nil {
				return err
			}
			cfg := loaded.Config
			ctx, cancel := context.WithTimeout(cmd.Context(), a.timeout)
			defer cancel()
			targetURL := args[0]

			content, err := tavily.New(cfg, a.timeout).Extract(ctx, targetURL)
			provider := "tavily"
			if err != nil {
				content, err = firecrawl.New(cfg, a.timeout).Scrape(ctx, targetURL, cfg.RetryMaxAttempts)
				provider = "firecrawl"
			}
			if err != nil {
				return err
			}
			out := fetchOutput{URL: targetURL, Provider: provider, Content: content}
			if a.format == "json" {
				return printJSON(out)
			}
			fmt.Print(content)
			if !strings.HasSuffix(content, "\n") {
				fmt.Println()
			}
			return nil
		},
	}
	return cmd
}

func (a *app) mapCommand() *cobra.Command {
	var instructions string
	var maxDepth, maxBreadth, limit, timeoutSeconds int
	cmd := &cobra.Command{
		Use:   "map URL",
		Short: "Map a website through Tavily Map",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			loaded, err := config.Load(a.configPath)
			if err != nil {
				return err
			}
			ctx, cancel := context.WithTimeout(cmd.Context(), time.Duration(timeoutSeconds+10)*time.Second)
			defer cancel()
			raw, err := tavily.New(loaded.Config, time.Duration(timeoutSeconds+10)*time.Second).Map(ctx, args[0], instructions, maxDepth, maxBreadth, limit, timeoutSeconds)
			if err != nil {
				return err
			}
			if a.format == "json" {
				var v any
				if err := json.Unmarshal(raw, &v); err == nil {
					return printJSON(v)
				}
			}
			fmt.Println(string(raw))
			return nil
		},
	}
	cmd.Flags().StringVar(&instructions, "instructions", "", "natural-language URL filter")
	cmd.Flags().IntVar(&maxDepth, "max-depth", 1, "maximum mapping depth")
	cmd.Flags().IntVar(&maxBreadth, "max-breadth", 20, "maximum links followed per page")
	cmd.Flags().IntVar(&limit, "limit", 50, "total URL processing limit")
	cmd.Flags().IntVar(&timeoutSeconds, "map-timeout", 150, "Tavily map timeout in seconds")
	return cmd
}

func (a *app) configCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "config", Short: "Inspect and update configuration"}
	cmd.AddCommand(&cobra.Command{
		Use:   "get",
		Short: "Print effective configuration with API keys masked",
		RunE: func(cmd *cobra.Command, args []string) error {
			loaded, err := config.Load(a.configPath)
			if err != nil {
				return err
			}
			return printJSON(maskedConfig(loaded))
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "check",
		Short: "Print config and test the Grok /models endpoint",
		RunE: func(cmd *cobra.Command, args []string) error {
			loaded, err := config.Load(a.configPath)
			if err != nil {
				return err
			}
			out := maskedConfig(loaded)
			ctx, cancel := context.WithTimeout(cmd.Context(), 10*time.Second)
			defer cancel()
			start := time.Now()
			models, err := grok.New(loaded.Config, 10*time.Second).Models(ctx)
			test := map[string]any{"status": "ok", "response_time_ms": time.Since(start).Milliseconds(), "models": models}
			if err != nil {
				test = map[string]any{"status": "error", "message": err.Error(), "response_time_ms": time.Since(start).Milliseconds()}
			}
			out["connection_test"] = test
			return printJSON(out)
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "set-model MODEL",
		Short: "Persist the default Grok model",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := a.configPath
			if path == "" {
				path = config.DefaultPath()
			}
			if err := config.SaveModel(path, args[0]); err != nil {
				return err
			}
			return printJSON(map[string]any{"status": "ok", "config_file": path, "model": args[0]})
		},
	})
	return cmd
}

func (a *app) modelsCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "models",
		Short: "List available Grok-compatible models",
		RunE: func(cmd *cobra.Command, args []string) error {
			loaded, err := config.Load(a.configPath)
			if err != nil {
				return err
			}
			ctx, cancel := context.WithTimeout(cmd.Context(), 10*time.Second)
			defer cancel()
			models, err := grok.New(loaded.Config, 10*time.Second).Models(ctx)
			if err != nil {
				return err
			}
			if a.format == "json" {
				return printJSON(map[string]any{"models": models})
			}
			for _, model := range models {
				fmt.Println(model)
			}
			return nil
		},
	}
}

func (a *app) lastCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "last",
		Short: "Print the last cached search result",
		RunE: func(cmd *cobra.Command, args []string) error {
			b, err := os.ReadFile(lastPath())
			if err != nil {
				if errors.Is(err, os.ErrNotExist) {
					return errors.New("没有找到最近一次搜索缓存")
				}
				return err
			}
			if a.format == "json" {
				fmt.Println(string(b))
				return nil
			}
			var out searchOutput
			if err := json.Unmarshal(b, &out); err != nil {
				return err
			}
			fmt.Println(out.Content)
			printSources(out.Sources)
			return nil
		},
	}
}

func collectExtraSources(ctx context.Context, cfg config.Config, query string, count int, timeout time.Duration) ([]sources.Source, error) {
	if count <= 0 {
		return nil, nil
	}
	var out []sources.Source
	tavilyCount := count
	firecrawlCount := 0
	if cfg.TavilyAPIKey != "" && cfg.FirecrawlAPIKey != "" {
		tavilyCount = count / 2
		firecrawlCount = count - tavilyCount
	} else if cfg.FirecrawlAPIKey != "" {
		tavilyCount = 0
		firecrawlCount = count
	}

	if tavilyCount > 0 {
		results, err := tavily.New(cfg, timeout).Search(ctx, query, tavilyCount)
		if err == nil {
			out = append(out, sources.FromTavily(results)...)
		}
	}
	if firecrawlCount > 0 {
		results, err := firecrawl.New(cfg, timeout).Search(ctx, query, firecrawlCount)
		if err == nil {
			out = append(out, sources.FromFirecrawl(results)...)
		}
	}
	return out, nil
}

func printJSON(v any) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(b))
	return nil
}

func printSources(srcs []sources.Source) {
	if len(srcs) == 0 {
		return
	}
	fmt.Println()
	fmt.Println("Sources:")
	for i, src := range srcs {
		if src.Title != "" {
			fmt.Printf("%d. %s - %s\n", i+1, src.Title, src.URL)
		} else {
			fmt.Printf("%d. %s\n", i+1, src.URL)
		}
	}
}

func writeSourcesFile(path string, srcs []sources.Source) error {
	b, err := sources.WriteJSON(srcs)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil && filepath.Dir(path) != "." {
		return err
	}
	return os.WriteFile(path, append(b, '\n'), 0o600)
}

func writeLast(out searchOutput) error {
	b, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return err
	}
	path := lastPath()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, append(b, '\n'), 0o600)
}

func lastPath() string {
	if dir, err := os.UserCacheDir(); err == nil && dir != "" {
		return filepath.Join(dir, "grok-search", "last-search.json")
	}
	return filepath.Join(".grok-search", "last-search.json")
}

func maskedConfig(loaded config.Loaded) map[string]any {
	cfg := loaded.Config
	return map[string]any{
		"GUDA_API_KEY":            config.Mask(cfg.GuDaAPIKey),
		"GUDA_BASE_URL":           valueOrUnset(cfg.GuDaBaseURL),
		"config_file":             loaded.Path,
		"GROK_API_URL":            valueOrUnset(cfg.GrokAPIURL),
		"GROK_API_KEY":            config.Mask(cfg.GrokAPIKey),
		"GROK_MODEL":              cfg.GrokModel,
		"TAVILY_ENABLED":          cfg.TavilyEnabled,
		"TAVILY_API_URL":          valueOrUnset(cfg.TavilyAPIURL),
		"TAVILY_API_KEY":          config.Mask(cfg.TavilyAPIKey),
		"FIRECRAWL_API_URL":       valueOrUnset(cfg.FirecrawlAPIURL),
		"FIRECRAWL_API_KEY":       config.Mask(cfg.FirecrawlAPIKey),
		"GROK_DEBUG":              cfg.Debug,
		"GROK_RETRY_MAX_ATTEMPTS": cfg.RetryMaxAttempts,
		"GROK_RETRY_MULTIPLIER":   cfg.RetryMultiplier,
		"GROK_RETRY_MAX_WAIT":     cfg.RetryMaxWait,
	}
}

func valueOrUnset(value string) string {
	if value == "" {
		return "未配置"
	}
	return value
}
