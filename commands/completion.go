// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package commands

import (
	"os"

	"github.com/spf13/cobra"
)

var CompletionCmd = &cobra.Command{
	Use:   "completion",
	Short: "Generates autocompletion scripts for bash and zsh",
}

var BashCmd = &cobra.Command{
	Use:   "bash",
	Short: "Generates the bash autocompletion scripts",
	Long: `To load completion, run

. <(mmctl completion bash)

To configure your bash shell to load completions for each session, add the above line to your ~/.bashrc
`,
	Run: bashCmdF,
}

var ZshCmd = &cobra.Command{
	Use:   "zsh",
	Short: "Generates the zsh autocompletion scripts",
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

func bashCmdF(cmd *cobra.Command, args []string) {
	_ = RootCmd.GenBashCompletion(os.Stdout)
}

func zshCmdF(cmd *cobra.Command, args []string) {
	zshInitialization := `
__mmctl_bash_source() {
	alias shopt=':'
	alias _expand=_bash_expand
	alias _complete=_bash_comp
	emulate -L sh
	setopt kshglob noshglob braceexpand
	source "$@"
}
__mmctl_type() {
	# -t is not supported by zsh
	if [ "$1" == "-t" ]; then
		shift
		# fake Bash 4 to disable "complete -o nospace". Instead
		# "compopt +-o nospace" is used in the code to toggle trailing
		# spaces. We don't support that, but leave trailing spaces on
		# all the time
		if [ "$1" = "__mmctl_compopt" ]; then
			echo builtin
			return 0
		fi
	fi
	type "$@"
}
__mmctl_compgen() {
	local completions w
	completions=( $(compgen "$@") ) || return $?
	# filter by given word as prefix
	while [[ "$1" = -* && "$1" != -- ]]; do
		shift
		shift
	done
	if [[ "$1" == -- ]]; then
		shift
	fi
	for w in "${completions[@]}"; do
		if [[ "${w}" = "$1"* ]]; then
			echo "${w}"
		fi
	done
}
__mmctl_compopt() {
	true # don't do anything. Not supported by bashcompinit in zsh
}
__mmctl_declare() {
	if [ "$1" == "-F" ]; then
		whence -w "$@"
	else
		builtin declare "$@"
	fi
}
__mmctl_ltrim_colon_completions()
{
	if [[ "$1" == *:* && "$COMP_WORDBREAKS" == *:* ]]; then
		# Remove colon-word prefix from COMPREPLY items
		local colon_word=${1%${1##*:}}
		local i=${#COMPREPLY[*]}
		while [[ $((--i)) -ge 0 ]]; do
			COMPREPLY[$i]=${COMPREPLY[$i]#"$colon_word"}
		done
	fi
}
__mmctl_get_comp_words_by_ref() {
	cur="${COMP_WORDS[COMP_CWORD]}"
	prev="${COMP_WORDS[${COMP_CWORD}-1]}"
	words=("${COMP_WORDS[@]}")
	cword=("${COMP_CWORD[@]}")
}
__mmctl_filedir() {
	local RET OLD_IFS w qw
	__debug "_filedir $@ cur=$cur"
	if [[ "$1" = \~* ]]; then
		# somehow does not work. Maybe, zsh does not call this at all
		eval echo "$1"
		return 0
	fi
	OLD_IFS="$IFS"
	IFS=$'\n'
	if [ "$1" = "-d" ]; then
		shift
		RET=( $(compgen -d) )
	else
		RET=( $(compgen -f) )
	fi
	IFS="$OLD_IFS"
	IFS="," __debug "RET=${RET[@]} len=${#RET[@]}"
	for w in ${RET[@]}; do
		if [[ ! "${w}" = "${cur}"* ]]; then
			continue
		fi
		if eval "[[ \"\${w}\" = *.$1 || -d \"\${w}\" ]]"; then
			qw="$(__mmctl_quote "${w}")"
			if [ -d "${w}" ]; then
				COMPREPLY+=("${qw}/")
			else
				COMPREPLY+=("${qw}")
			fi
		fi
	done
}
__mmctl_quote() {
    if [[ $1 == \'* || $1 == \"* ]]; then
        # Leave out first character
        printf %q "${1:1}"
    else
    	printf %q "$1"
    fi
}
autoload -U +X bashcompinit && bashcompinit
# use word boundary patterns for BSD or GNU sed
LWORD='[[:<:]]'
RWORD='[[:>:]]'
if sed --help 2>&1 | grep -q GNU; then
	LWORD='\<'
	RWORD='\>'
fi
__mmctl_convert_bash_to_zsh() {
	sed \
	-e 's/declare -F/whence -w/' \
	-e 's/local \([a-zA-Z0-9_]*\)=/local \1; \1=/' \
	-e 's/flags+=("\(--.*\)=")/flags+=("\1"); two_word_flags+=("\1")/' \
	-e 's/must_have_one_flag+=("\(--.*\)=")/must_have_one_flag+=("\1")/' \
	-e "s/${LWORD}_filedir${RWORD}/__mmctl_filedir/g" \
	-e "s/${LWORD}_get_comp_words_by_ref${RWORD}/__mmctl_get_comp_words_by_ref/g" \
	-e "s/${LWORD}__ltrim_colon_completions${RWORD}/__mmctl_ltrim_colon_completions/g" \
	-e "s/${LWORD}compgen${RWORD}/__mmctl_compgen/g" \
	-e "s/${LWORD}compopt${RWORD}/__mmctl_compopt/g" \
	-e "s/${LWORD}declare${RWORD}/__mmctl_declare/g" \
	-e "s/\\\$(type${RWORD}/\$(__mmctl_type/g" \
	<<'BASH_COMPLETION_EOF'
`

	zshTail := `
BASH_COMPLETION_EOF
}
__mmctl_bash_source <(__mmctl_convert_bash_to_zsh)
`

	os.Stdout.Write([]byte(zshInitialization))
	_ = RootCmd.GenBashCompletion(os.Stdout)
	os.Stdout.Write([]byte(zshTail))
}
