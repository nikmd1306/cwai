package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/nikmd1306/cwai/internal/ai"
	"github.com/nikmd1306/cwai/internal/config"
	"github.com/nikmd1306/cwai/internal/diff"
	"github.com/nikmd1306/cwai/internal/git"
	"github.com/nikmd1306/cwai/internal/prompt"
	"github.com/nikmd1306/cwai/internal/update"
	"github.com/spf13/cobra"
)

var hookFlag bool
var yesFlag bool

var rootCmd = &cobra.Command{
	Use:   "cwai",
	Short: "AI-powered commit message generator",
	Long:  "cwai generates conventional commit messages using AI from your staged changes.",
	Args:  cobra.ArbitraryArgs,
	RunE:  runRoot,
}

func init() {
	rootCmd.Flags().BoolVar(&hookFlag, "hook", false, "run in git hook mode (prepare-commit-msg)")
	_ = rootCmd.Flags().MarkHidden("hook")
	rootCmd.Flags().BoolVarP(&yesFlag, "yes", "y", false, "auto-accept generated commit message")

	rootCmd.CompletionOptions.DisableDefaultCmd = true

	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(setupCmd)
	rootCmd.AddCommand(hookCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(updateCmd)
	rootCmd.AddCommand(changelogCmd)
	rootCmd.AddCommand(completionCmd)
}

func Execute() error {
	type updateResult struct {
		info *update.ReleaseInfo
	}

	var updateCh chan updateResult
	if os.Getenv("CWAI_NO_UPDATE_NOTIFIER") == "" {
		updateCh = make(chan updateResult, 1)
		go func() {
			info, _ := update.CheckForUpdate(Version)
			updateCh <- updateResult{info: info}
		}()
	}

	err := rootCmd.Execute()

	if updateCh != nil {
		select {
		case result := <-updateCh:
			if result.info != nil {
				fmt.Fprintf(os.Stderr, "\nA new version of cwai is available: %s → %s\n", Version, result.info.Version)
				fmt.Fprintf(os.Stderr, "Run 'cwai update' to upgrade\n")
			}
		case <-time.After(2 * time.Second):
		}
	}

	return err
}

func runRoot(cmd *cobra.Command, args []string) error {
	if hookFlag {
		return runHookMode(args)
	}
	return runStandalone()
}

func runHookMode(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("hook mode requires commit message file path")
	}

	msgFile := args[0]
	source := ""
	if len(args) > 1 {
		source = args[1]
	}

	if source == "merge" || source == "squash" || source == "commit" {
		return nil
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}
	if err := cfg.Validate(); err != nil {
		return err
	}

	message, err := generate(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cwai: %v\n", err)
		return nil
	}

	return os.WriteFile(msgFile, []byte(message), 0o644)
}

func runStandalone() error {
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

	files, err := git.StagedFiles()
	if err != nil {
		return err
	}
	if len(files) == 0 {
		return fmt.Errorf("no staged changes. Use 'git add' first")
	}

	fmt.Printf("Staged files (%d):\n", len(files))
	for _, f := range files {
		fmt.Printf("  %s\n", f)
	}
	fmt.Println()

	message, err := generate(cfg)
	if err != nil {
		return err
	}

	if yesFlag {
		fmt.Println(message)
		if err := git.Commit(message); err != nil {
			return fmt.Errorf("commit failed: %w", err)
		}
		fmt.Println("Committed successfully!")
		return nil
	}

	for {
		fmt.Println("Generated commit message:")
		fmt.Printf("\n  %s\n\n", strings.ReplaceAll(message, "\n", "\n  "))
		fmt.Print("[y]es / [e]dit / [r]egenerate / [n]o: ")

		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(strings.ToLower(input))

		switch input {
		case "y", "yes", "":
			if err := git.Commit(message); err != nil {
				return fmt.Errorf("commit failed: %w", err)
			}
			fmt.Println("Committed successfully!")
			return nil

		case "e", "edit":
			fmt.Print("Edit commit message (press Enter to keep current):\n> ")
			edited, _ := reader.ReadString('\n')
			edited = strings.TrimSpace(edited)
			if edited != "" {
				message = edited
			}

		case "r", "regenerate":
			fmt.Println("Regenerating...")
			message, err = generate(cfg)
			if err != nil {
				return err
			}

		case "n", "no":
			fmt.Println("Aborted.")
			return nil

		default:
			fmt.Println("Invalid option. Please choose y/e/r/n.")
		}
	}
}

func generate(cfg *config.Config) (string, error) {
	diffText, err := diff.Truncate(cfg.MaxTokensInput)
	if err != nil {
		return "", err
	}

	if !git.HasCommits() {
		diffText = "[CONTEXT: This is the initial commit in a new repository. For initial repository setup, use type \"chore\" and scope \"project\".]\n\n" + diffText
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
		Verbosity:          cfg.Verbosity,
		StructuredOutput:   cfg.StructuredOutput,
	})

	isStructured := client.IsStructuredOutput()
	messages := prompt.BuildMessages(cfg.Language, diffText, isStructured)

	return client.GenerateCommitMessage(messages)
}
