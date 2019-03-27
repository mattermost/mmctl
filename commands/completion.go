package commands

import (
	"os"

	"github.com/spf13/cobra"
)

var CompletionCmd = &cobra.Command{
	Use: "completion",
	Short: "Generates autocompletion scripts for bash and zsh",
}

var BashCmd = &cobra.Command{
	Use: "bash",
	Short: "Generates the bash autocompletion scripts",
	Long: `To load completion, run

. <(mmctl completion bash)

To configure your bash shell to load completions for each session, add the above line to your ~/.bashrc
`,
	Run: bashCmdF,
}

var ZshCmd = &cobra.Command{
	Use: "zsh",
	Short: "Generates the zsh autocompletion scripts (EXPERIMENTAL)",
	Long: `To load completion, run

. <(mmctl completion zsh)

To configure your zsh shell to load completions for each session, add the above line to your ~/.zshrc
`,
	Run: zshCmdF,
}

func init() {
	CompletionCmd.AddCommand(
		BashCmd,
		ZshCmd,
	)

	RootCmd.AddCommand(CompletionCmd)
}

func bashCmdF(command *cobra.Command, args []string) {
	RootCmd.GenBashCompletion(os.Stdout)
}

func zshCmdF(command *cobra.Command, args []string) {
	RootCmd.GenZshCompletion(os.Stdout)
}
