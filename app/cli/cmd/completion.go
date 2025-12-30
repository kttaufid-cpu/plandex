package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate shell completion scripts",
	Long: `Generate shell completion scripts for Plandex.

To load completions:

Bash:
  $ source <(plandex completion bash)

  # To load completions for each session, execute once:
  # Linux:
  $ plandex completion bash > /etc/bash_completion.d/plandex
  # macOS:
  $ plandex completion bash > $(brew --prefix)/etc/bash_completion.d/plandex

Zsh:
  # If shell completion is not already enabled in your environment,
  # you will need to enable it. You can execute the following once:
  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  # To load completions for each session, execute once:
  $ plandex completion zsh > "${fpath[1]}/_plandex"

  # You will need to start a new shell for this setup to take effect.

Fish:
  $ plandex completion fish | source

  # To load completions for each session, execute once:
  $ plandex completion fish > ~/.config/fish/completions/plandex.fish

PowerShell:
  PS> plandex completion powershell | Out-String | Invoke-Expression

  # To load completions for every new session, run:
  PS> plandex completion powershell > plandex.ps1
  # and source this file from your PowerShell profile.
`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	Run: func(cmd *cobra.Command, args []string) {
		switch args[0] {
		case "bash":
			RootCmd.GenBashCompletion(os.Stdout)
		case "zsh":
			RootCmd.GenZshCompletion(os.Stdout)
		case "fish":
			RootCmd.GenFishCompletion(os.Stdout, true)
		case "powershell":
			RootCmd.GenPowerShellCompletionWithDesc(os.Stdout)
		}
	},
}

func init() {
	RootCmd.AddCommand(completionCmd)
}
