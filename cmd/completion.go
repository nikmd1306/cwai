package cmd

import (
	"fmt"
	"os"

	"github.com/nikmd1306/cwai/internal/completion"
	"github.com/spf13/cobra"
)

var completionShellFlag string

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish]",
	Short: "Generate or install shell completion scripts",
	Long: `Generate shell completion scripts for cwai.

To output the completion script to stdout:
  cwai completion bash
  cwai completion zsh
  cwai completion fish

To auto-install completions:
  cwai completion install [--shell bash|zsh|fish]

To remove installed completions:
  cwai completion uninstall [--shell bash|zsh|fish]`,
	ValidArgs: []string{"bash", "zsh", "fish"},
	Args:      cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	RunE: func(cmd *cobra.Command, args []string) error {
		switch args[0] {
		case "bash":
			return rootCmd.GenBashCompletionV2(os.Stdout, true)
		case "zsh":
			return rootCmd.GenZshCompletion(os.Stdout)
		case "fish":
			return rootCmd.GenFishCompletion(os.Stdout, true)
		}
		return nil
	},
}

var completionInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install shell completions for cwai",
	RunE: func(cmd *cobra.Command, args []string) error {
		sh, path, err := completion.Install(rootCmd, completionShellFlag)
		if err != nil {
			return err
		}
		fmt.Printf("Completion for %s installed to %s\n", sh, path)
		if hint := completion.PostInstallHint(sh); hint != "" {
			fmt.Println()
			fmt.Println(hint)
		}
		return nil
	},
}

var completionUninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Remove installed shell completions for cwai",
	RunE: func(cmd *cobra.Command, args []string) error {
		sh, path, err := completion.Uninstall(completionShellFlag)
		if err != nil {
			return err
		}
		fmt.Printf("Completion for %s removed from %s\n", sh, path)
		return nil
	},
}

func init() {
	completionInstallCmd.Flags().StringVar(&completionShellFlag, "shell", "", "shell type (bash, zsh, fish)")
	completionUninstallCmd.Flags().StringVar(&completionShellFlag, "shell", "", "shell type (bash, zsh, fish)")

	completionCmd.AddCommand(completionInstallCmd)
	completionCmd.AddCommand(completionUninstallCmd)
}
