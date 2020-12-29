package cmd

import (
	"bytes"
	"io"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var (
	completionShells = map[string]func(out io.Writer, cmd *cobra.Command) error{
		"bash": runCompletionBash,
		"zsh":  runCompletionZsh,
	}
)

// NewCmdCompletion creates the `completion` command
func NewCompletionCmd() *cobra.Command {
	var shells []string
	for s := range completionShells {
		shells = append(shells, s)
	}

	cmd := &cobra.Command{
		Use:                   "completion SHELL",
		DisableFlagsInUseLine: true,
		Short:                 "Output shell completion code for the specified shell (bash or zsh)",

		RunE: func(cmd *cobra.Command, args []string) error {
			return RunCompletion(os.Stdout, cmd, args)
		},
		ValidArgs: shells,
	}

	return cmd
}

// RunCompletion checks given arguments and executes command
func RunCompletion(out io.Writer, cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return errors.New("Shell not specified.")
	}
	if len(args) > 1 {
		return errors.New("Too many arguments. Expected only the shell type.")
	}
	run, found := completionShells[args[0]]
	if !found {
		return errors.Errorf("Unsupported shell type %q.", args[0])
	}

	return run(out, cmd.Parent())
}

func runCompletionBash(out io.Writer, cmd *cobra.Command) error {
	return cmd.GenBashCompletion(out)
}

func runCompletionZsh(out io.Writer, cmd *cobra.Command) error {
	zshHead := "#compdef knx-exporter\n"

	_, e := out.Write([]byte(zshHead))
	if e != nil {
		return e
	}

	zshInitialization := `
__knx-exporter_bash_source() {
	alias shopt=':'
	alias _expand=_bash_expand
	alias _complete=_bash_comp
	emulate -L sh
	setopt kshglob noshglob braceexpand

	source "$@"
}

__knx-exporter_type() {
	# -t is not supported by zsh
	if [ "$1" == "-t" ]; then
		shift

		# fake Bash 4 to disable "complete -o nospace". Instead
		# "compopt +-o nospace" is used in the code to toggle trailing
		# spaces. We don't support that, but leave trailing spaces on
		# all the time
		if [ "$1" = "__knx-exporter_compopt" ]; then
			echo builtin
			return 0
		fi
	fi
	type "$@"
}

__knx-exporter_compgen() {
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

__knx-exporter_compopt() {
	true # don't do anything. Not supported by bashcompinit in zsh
}

__knx-exporter_ltrim_colon_completions()
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

__knx-exporter_get_comp_words_by_ref() {
	cur="${COMP_WORDS[COMP_CWORD]}"
	prev="${COMP_WORDS[${COMP_CWORD}-1]}"
	words=("${COMP_WORDS[@]}")
	cword=("${COMP_CWORD[@]}")
}

__knx-exporter_filedir() {
	local RET OLD_IFS w qw

	__knx-exporter_debug "_filedir $@ cur=$cur"
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

	IFS="," __knx-exporter_debug "RET=${RET[@]} len=${#RET[@]}"

	for w in ${RET[@]}; do
		if [[ ! "${w}" = "${cur}"* ]]; then
			continue
		fi
		if eval "[[ \"\${w}\" = *.$1 || -d \"\${w}\" ]]"; then
			qw="$(__knx-exporter_quote "${w}")"
			if [ -d "${w}" ]; then
				COMPREPLY+=("${qw}/")
			else
				COMPREPLY+=("${qw}")
			fi
		fi
	done
}

__knx-exporter_quote() {
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

__knx-exporter_convert_bash_to_zsh() {
	sed \
	-e 's/declare -F/whence -w/' \
	-e 's/_get_comp_words_by_ref "\$@"/_get_comp_words_by_ref "\$*"/' \
	-e 's/local \([a-zA-Z0-9_]*\)=/local \1; \1=/' \
	-e 's/flags+=("\(--.*\)=")/flags+=("\1"); two_word_flags+=("\1")/' \
	-e 's/must_have_one_flag+=("\(--.*\)=")/must_have_one_flag+=("\1")/' \
	-e "s/${LWORD}_filedir${RWORD}/__knx-exporter_filedir/g" \
	-e "s/${LWORD}_get_comp_words_by_ref${RWORD}/__knx-exporter_get_comp_words_by_ref/g" \
	-e "s/${LWORD}__ltrim_colon_completions${RWORD}/__knx-exporter_ltrim_colon_completions/g" \
	-e "s/${LWORD}compgen${RWORD}/__knx-exporter_compgen/g" \
	-e "s/${LWORD}compopt${RWORD}/__knx-exporter_compopt/g" \
	-e "s/${LWORD}declare${RWORD}/builtin declare/g" \
	-e "s/\\\$(type${RWORD}/\$(__knx-exporter_type/g" \
	<<'BASH_COMPLETION_EOF'
`
	_, _ = out.Write([]byte(zshInitialization))
	buf := new(bytes.Buffer)
	e = cmd.GenBashCompletion(buf)
	if e != nil {
		return e
	}
	_, _ = out.Write(buf.Bytes())

	zshTail := `
BASH_COMPLETION_EOF
}

__knx-exporter_bash_source <(__knx-exporter_convert_bash_to_zsh)
_complete knx-exporter 2>/dev/null
`
	_, _ = out.Write([]byte(zshTail))
	return nil
}
