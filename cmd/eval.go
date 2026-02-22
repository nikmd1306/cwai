//go:build eval

package cmd

import (
	"fmt"
	"time"

	"github.com/nikmd1306/cwai/internal/config"
	"github.com/nikmd1306/cwai/internal/eval"
	"github.com/spf13/cobra"
)

var evalCmd = &cobra.Command{
	Use:   "eval",
	Short: "Run evaluation against golden dataset",
	RunE:  runEval,
}

var evalCompareCmd = &cobra.Command{
	Use:   "compare <run-a> <run-b>",
	Short: "Compare two evaluation runs",
	Args:  cobra.ExactArgs(2),
	RunE:  runEvalCompare,
}

var (
	evalDataset          string
	evalOutputDir        string
	evalTags             []string
	evalModel            string
	evalAPIURL           string
	evalStructuredOutput string
	evalDelay            time.Duration
)

func init() {
	evalCmd.Flags().StringVar(&evalDataset, "dataset", "experiments/dataset/golden.jsonl", "path to dataset file")
	evalCmd.Flags().StringVar(&evalOutputDir, "out", "experiments/runs", "output directory for run results")
	evalCmd.Flags().StringSliceVar(&evalTags, "tags", nil, "filter samples by tags")
	evalCmd.Flags().StringVar(&evalModel, "model", "", "model to use (overrides config)")
	evalCmd.Flags().StringVar(&evalAPIURL, "api-url", "", "API URL (overrides config)")
	evalCmd.Flags().StringVar(&evalStructuredOutput, "structured-output", "", "structured output mode (overrides config)")
	evalCmd.Flags().DurationVar(&evalDelay, "delay", 0, "delay between API calls")

	evalCmd.AddCommand(evalCompareCmd)
}

func runEval(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	if err := cfg.Validate(); err != nil {
		return err
	}

	model := cfg.Model
	if evalModel != "" {
		model = evalModel
	}
	apiURL := cfg.APIURL
	if evalAPIURL != "" {
		apiURL = evalAPIURL
	}
	structuredOutput := cfg.StructuredOutput
	if evalStructuredOutput != "" {
		structuredOutput = evalStructuredOutput
	}

	runID := time.Now().UTC().Format("2006-01-02T15-04-05") + "_" + model

	runCfg := eval.RunConfig{
		RunID:            runID,
		DatasetPath:      evalDataset,
		Model:            model,
		APIURL:           apiURL,
		APIKey:           cfg.APIKey,
		MaxTokensOutput:  cfg.MaxTokensOutput,
		Temperature:      cfg.Temperature,
		HasTemperature:   cfg.HasTemperature,
		ReasoningEffort:  cfg.ReasoningEffort,
		StructuredOutput: structuredOutput,
		Language:         cfg.Language,
		Tags:             evalTags,
		OutputDir:        evalOutputDir,
		Delay:            evalDelay,
	}

	fmt.Printf("Starting evaluation run: %s\n", runID)
	fmt.Printf("Dataset: %s\n", evalDataset)
	fmt.Printf("Model: %s\n", model)
	fmt.Println()

	_, err = eval.Run(runCfg)
	return err
}

func runEvalCompare(cmd *cobra.Command, args []string) error {
	result, err := eval.CompareRuns(args[0], args[1])
	if err != nil {
		return err
	}
	fmt.Println(result)
	return nil
}
