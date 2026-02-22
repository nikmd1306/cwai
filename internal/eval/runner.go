//go:build eval

package eval

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/nikmd1306/cwai/internal/ai"
	"github.com/nikmd1306/cwai/internal/prompt"
)

type RunConfig struct {
	RunID              string        `json:"run_id"`
	DatasetPath        string        `json:"dataset_path"`
	Model              string        `json:"model"`
	APIURL             string        `json:"api_url"`
	APIKey             string        `json:"-"`
	MaxTokensOutput    int           `json:"max_tokens_output"`
	HasMaxTokensOutput bool          `json:"has_max_tokens_output"`
	Temperature        float64       `json:"temperature"`
	HasTemperature     bool          `json:"has_temperature"`
	ReasoningEffort    string        `json:"reasoning_effort,omitempty"`
	StructuredOutput   string        `json:"structured_output"`
	Language           string        `json:"language"`
	Tags               []string      `json:"tags,omitempty"`
	OutputDir          string        `json:"output_dir"`
	Delay              time.Duration `json:"-"`
}

func Run(cfg RunConfig) ([]SampleResult, error) {
	samples, err := LoadDatasetFiltered(cfg.DatasetPath, cfg.Tags)
	if err != nil {
		return nil, fmt.Errorf("load dataset: %w", err)
	}
	if len(samples) == 0 {
		return nil, fmt.Errorf("no samples to evaluate")
	}

	client := ai.NewClient(ai.Params{
		APIKey:             cfg.APIKey,
		APIURL:             cfg.APIURL,
		Model:              cfg.Model,
		MaxTokensOutput:    cfg.MaxTokensOutput,
		HasMaxTokensOutput: cfg.HasMaxTokensOutput,
		Temperature:        cfg.Temperature,
		HasTemperature:     cfg.HasTemperature,
		ReasoningEffort:    cfg.ReasoningEffort,
		StructuredOutput:   cfg.StructuredOutput,
	})

	isStructured := client.IsStructuredOutput()
	language := cfg.Language
	if language == "" {
		language = "en"
	}

	runDir := filepath.Join(cfg.OutputDir, cfg.RunID)
	if err := os.MkdirAll(runDir, 0o755); err != nil {
		return nil, fmt.Errorf("create run directory: %w", err)
	}

	configData, _ := json.MarshalIndent(cfg, "", "  ")
	if err := os.WriteFile(filepath.Join(runDir, "config.json"), configData, 0o644); err != nil {
		return nil, fmt.Errorf("save config: %w", err)
	}

	resultsFile, err := os.Create(filepath.Join(runDir, "results.jsonl"))
	if err != nil {
		return nil, fmt.Errorf("create results file: %w", err)
	}
	defer func() { _ = resultsFile.Close() }()

	var results []SampleResult
	for i, sample := range samples {
		if i > 0 && cfg.Delay > 0 {
			time.Sleep(cfg.Delay)
		}

		result := evaluateSample(client, sample, language, isStructured)
		results = append(results, result)

		line, _ := json.Marshal(result)
		if _, err := fmt.Fprintf(resultsFile, "%s\n", line); err != nil {
			log.Printf("WARNING: failed to write result for sample %s: %v", sample.ID, err)
		}

		fmt.Printf("  [%d/%d] %s: ", i+1, len(samples), sample.ID)
		if result.Error != "" {
			fmt.Printf("ERROR: %s\n", result.Error)
		} else {
			fmt.Printf("type=%v scope=%v format=%v sim=%.2f latency=%dms\n",
				result.TypeMatch, result.ScopeMatch, result.FormatValid,
				result.DescSimilarity, result.LatencyMs)
		}
	}

	summary := Summarize(cfg.RunID, cfg.Model, cfg.APIURL, isStructured, results)
	summary.Timestamp = time.Now().UTC().Format(time.RFC3339)
	summaryData, _ := json.MarshalIndent(summary, "", "  ")
	if err := os.WriteFile(filepath.Join(runDir, "summary.json"), summaryData, 0o644); err != nil {
		return nil, fmt.Errorf("save summary: %w", err)
	}

	fmt.Printf("\nRun %s complete. Summary saved to %s\n", cfg.RunID, filepath.Join(runDir, "summary.json"))
	fmt.Println(FormatSummary(summary))

	return results, nil
}

func evaluateSample(client *ai.Client, sample Sample, language string, structured bool) SampleResult {
	result := SampleResult{
		SampleID:        sample.ID,
		ExpectedMessage: sample.ExpectedMessage,
	}

	messages := prompt.BuildMessages(language, sample.Diff, structured)

	start := time.Now()
	generated, err := client.GenerateCommitMessage(messages)
	result.LatencyMs = time.Since(start).Milliseconds()

	if err != nil {
		result.Error = err.Error()
		return result
	}

	result.GeneratedMessage = generated
	if tokens := client.LastTotalTokens(); tokens > 0 {
		result.TokensUsed = tokens
	} else {
		result.TokensUsed = len(sample.Diff) / 4
	}
	result.FormatValid = ValidateFormat(generated)

	genType, genScope, genDesc, parseErr := ParseCommitHeader(generated)
	if parseErr == nil {
		result.TypeMatch = genType == sample.ExpectedType
		result.ScopeMatch = genScope == sample.ExpectedScope

		_, _, expectedDesc, expectedParseErr := ParseCommitHeader(sample.ExpectedMessage)
		if expectedParseErr != nil {
			log.Printf("WARNING: malformed expected message for sample %s: %v", sample.ID, expectedParseErr)
		} else {
			result.DescSimilarity = computeHybridOrFallback(client, genDesc, expectedDesc)
		}
	}

	return result
}

func computeHybridOrFallback(client *ai.Client, genDesc, expectedDesc string) float64 {
	embeddings, err := client.GetEmbeddings([]string{genDesc, expectedDesc})
	if err != nil || len(embeddings) < 2 {
		return ComputeDescSimilarity(genDesc, expectedDesc)
	}
	cosine := CosineSimilarity(embeddings[0], embeddings[1])
	jw := JaroWinkler(genDesc, expectedDesc)
	return HybridSimilarity(cosine, jw)
}
