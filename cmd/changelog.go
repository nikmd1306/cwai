package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/nikmd1306/cwai/internal/ai"
	"github.com/nikmd1306/cwai/internal/config"
	"github.com/nikmd1306/cwai/internal/git"
	"github.com/nikmd1306/cwai/internal/prompt"
	"github.com/spf13/cobra"
)

var (
	changelogFrom   string
	changelogTo     string
	changelogOutput string
)

var changelogCmd = &cobra.Command{
	Use:   "changelog",
	Short: "Generate AI-powered release notes from commits",
	Long:  "Generates human-readable release notes in Markdown by analyzing commits between two git references using AI.",
	RunE:  runChangelog,
}

func init() {
	changelogCmd.Flags().StringVar(&changelogFrom, "from", "", "start reference (default: latest tag)")
	changelogCmd.Flags().StringVar(&changelogTo, "to", "HEAD", "end reference")
	changelogCmd.Flags().StringVarP(&changelogOutput, "output", "o", "", "write output to file instead of stdout")
}

func runChangelog(cmd *cobra.Command, args []string) error {
	if !git.IsRepo() {
		return fmt.Errorf("not a git repository")
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}
	if err := cfg.Validate(); err != nil {
		return err
	}

	from := changelogFrom
	if from == "" {
		tag, err := git.GetLatestTag()
		if err != nil {
			return fmt.Errorf("cannot determine latest tag: %w", err)
		}
		from = tag
	}

	commits, err := git.GetCommitsBetween(from, changelogTo)
	if err != nil {
		return fmt.Errorf("cannot get commits: %w", err)
	}
	if len(commits) == 0 {
		return fmt.Errorf("no commits found between %s and %s", from, changelogTo)
	}

	commitsText := strings.Join(commits, "\n")

	client := ai.NewClient(ai.Params{
		APIKey:             cfg.APIKey,
		APIURL:             cfg.APIURL,
		Model:              cfg.Model,
		MaxTokensOutput:    cfg.MaxTokensOutput,
		HasMaxTokensOutput: cfg.HasMaxTokensOutput,
		Temperature:        cfg.Temperature,
		HasTemperature:     cfg.HasTemperature,
		ReasoningEffort:    cfg.ReasoningEffort,
		Verbosity:          cfg.Verbosity,
	})

	messages := prompt.BuildChangelogMessages(cfg.Language, commitsText)

	result, err := client.GenerateText(messages)
	if err != nil {
		fmt.Fprintf(os.Stderr, "AI generation failed, using fallback grouping: %v\n", err)
		result = prompt.GroupCommitsByType(commits)
	}

	if changelogOutput != "" {
		if err := os.WriteFile(changelogOutput, []byte(result+"\n"), 0o644); err != nil {
			return fmt.Errorf("write output file: %w", err)
		}
		fmt.Fprintf(os.Stderr, "Release notes written to %s\n", changelogOutput)
		return nil
	}

	fmt.Println(result)
	return nil
}
