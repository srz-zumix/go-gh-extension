package completion

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cli/cli/v2/pkg/cmdutil"
	"github.com/spf13/cobra"
)

// ShellTypes contains supported shell types
var ShellTypes = []string{"bash", "zsh", "fish", "powershell"}

// GetExecutableName returns the base name of the executable
func GetExecutableName() string {
	execPath, err := os.Executable()
	if err != nil {
		return ""
	}
	return filepath.Base(execPath)
}

// GetExtensionName returns the extension name without 'gh-' prefix
// e.g., "gh-team-kit" -> "team-kit", "gh-label-kit" -> "label-kit"
func GetExtensionName() string {
	execName := GetExecutableName()
	if strings.HasPrefix(execName, "gh-") {
		return execName[3:]
	}
	return execName
}

// IsGhExtension checks if the command is running as a gh extension
func IsGhExtension() bool {
	execPath, err := os.Executable()
	if err == nil {
		absPath, err := filepath.Abs(execPath)
		if err == nil {
			execName := GetExecutableName()
			return strings.Contains(absPath, "extensions/"+execName)
		}
	}
	return false
}

// NewCompletionCmd creates a new completion command
func NewCompletionCmd() *cobra.Command {
	var shell string

	cmd := &cobra.Command{
		Use:   "completion -s <shell>",
		Short: "Generate shell completion script",
		Long: fmt.Sprintf(`Generate shell completion script for %s.

Automatically detects the calling context and generates the appropriate completion:
- When called as 'gh %s': Generates a patch for gh completion
- When called as '%s': Generates standard completion script
`, GetExecutableName(), GetExtensionName(), GetExecutableName()),
		DisableFlagsInUseLine: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Auto-detect if running as gh extension
			if IsGhExtension() {
				return GenerateExtensionCompletionPatch(shell)
			}

			// Use executable name (with hyphen) to avoid conflicts with gh CLI completion
			execName := GetExecutableName()
			root := cmd.Root()
			originalUse := root.Use
			root.Use = execName
			defer func() {
				root.Use = originalUse
			}()

			var buf bytes.Buffer
			var err error

			switch shell {
			case "bash":
				err = root.GenBashCompletion(&buf)
			case "zsh":
				err = root.GenZshCompletion(&buf)
			case "fish":
				err = root.GenFishCompletion(&buf, true)
			case "powershell":
				err = root.GenPowerShellCompletionWithDesc(&buf)
			}

			if err != nil {
				return err
			}

			// Replace hyphens with underscores for function names (shell compatibility)
			output := buf.String()
			output = strings.ReplaceAll(output, execName, strings.ReplaceAll(execName, "-", "_"))

			_, err = os.Stdout.WriteString(output)
			return err
		},
	}

	cmdutil.StringEnumFlag(cmd, &shell, "shell", "s", "", ShellTypes, "Shell type")
	cmd.MarkFlagRequired("shell")

	return cmd
}

// GenerateExtensionCompletionPatch generates a completion patch script for gh extensions
func GenerateExtensionCompletionPatch(shell string) error {
	var script string
	extName := GetExtensionName()                      // e.g., "team-kit"
	execName := GetExecutableName()                    // e.g., "gh-team-kit"
	funcName := strings.ReplaceAll(extName, "-", "_") // e.g., "team_kit"

	switch shell {
	case "bash":
		script = generateBashPatch(execName, extName)
	case "zsh":
		script = generateZshPatch(execName, extName, funcName)
	case "fish":
		script = generateFishPatch(execName, extName)
	case "powershell":
		script = generatePowerShellPatch(execName, extName)
	default:
		return nil
	}

	_, err := os.Stdout.WriteString(script)
	return err
}

func generateBashPatch(execName, extName string) string {
	return fmt.Sprintf(`# %s extension completion patch for bash
# Source this file after gh completion to enable 'gh %s' completion

# Override __gh_get_completion_results to handle %s
if declare -f __gh_get_completion_results >/dev/null 2>&1; then
    eval "$(declare -f __gh_get_completion_results | sed '1s/.*/__gh_get_completion_results_original()/')"

    __gh_get_completion_results() {
        local requestComp lastParam lastChar args

        # Check for redirections or pipes - skip completion to avoid executing redirections
        local i
        for (( i=0; i<${#words[@]}; i++ )); do
            case "${words[i]}" in
                ">"| ">>" | "<" | "|" | "2>" | "2>>" | "&>" | "&>>")
                    # Redirection or pipe detected, return without setting COMPREPLY
                    # This allows bash to use default file completion
                    return 1
                    ;;
            esac
        done

        # Prepare the command to request completions for the program.
        args=("${words[@]:1}")

        # Check if this is a %s completion
        if [[ "${args[0]}" == "%s" ]]; then
            # Route directly to the extension
            # Remove empty trailing arguments to avoid extra spaces
            local comp_args=()
            local i
            for (( i=1; i<${#args[@]}; i++ )); do
                comp_args+=("${args[i]}")
            done

            requestComp="gh %s __complete"
            for arg in "${comp_args[@]}"; do
                if [[ -n "$arg" ]]; then
                    requestComp="${requestComp} $arg"
                fi
            done
        else
            # Use the original logic
            requestComp="${words[0]} __complete ${args[*]}"
        fi

        lastParam=${words[$((${#words[@]}-1))]}
        lastChar=${lastParam:$((${#lastParam}-1)):1}
        __gh_debug "lastParam ${lastParam}, lastChar ${lastChar}"

        if [[ -z ${cur} && ${lastChar} != = ]]; then
            __gh_debug "Adding extra empty parameter"
            requestComp="${requestComp} ''"
        fi

        if [[ ${cur} == -*=* ]]; then
            cur="${cur#*=}"
        fi

        __gh_debug "Calling ${requestComp}"
        out=$(eval "${requestComp}" 2>/dev/null)

        directive=${out##*:}
        out=${out%%%%:*}
        if [[ ${directive} == "${out}" ]]; then
            directive=0
        fi
        __gh_debug "The completion directive is: ${directive}"
        __gh_debug "The completions are: ${out}"
    }
else
    echo "Warning: gh completion not loaded. Please source gh completion first." >&2
    echo "Run: gh completion -s bash | source" >&2
fi
`, execName, extName, extName, extName, extName, extName)
}

func generateZshPatch(execName, extName, funcName string) string {
	return fmt.Sprintf(`# %s extension completion patch for zsh
# Source this file after gh completion to enable 'gh %s' completion

# Save the original _gh function and redefine it
if (( $+functions[_gh] )); then
    # Rename original function
    functions[__gh_original_completion]=$functions[_gh]

    # Redefine _gh to intercept team-kit completion
    _gh() {
        # Check if this is a team-kit completion request
        if [[ ${#words[@]} -ge 2 && "${words[2]}" == "%s" ]]; then
            _gh_%s
            return
        fi

        # Otherwise, call the original function (zsh completion functions don't take arguments)
        __gh_original_completion
    }

    _gh_%s() {
        local -a completions
        local -a completions_with_descriptions
        local -a response

        # Build the command - skip 'gh' and 'team-kit', take the rest
        local requestComp="gh %s __complete"
        local i
        for (( i=3; i<=${#words[@]}; i++ )); do
            if [[ -n "${words[i]}" ]]; then
                requestComp="${requestComp} ${words[i]}"
            fi
        done

        # Add empty parameter if the cursor is at the end
        if [[ -z "${words[CURRENT]}" ]] || [[ "${words[CURRENT]}" == "" ]]; then
            requestComp="${requestComp} ''"
        fi

        # Execute completion
        response=(${(f)"$(eval ${requestComp} 2>/dev/null)"})

        # Parse response
        for line in $response; do
            if [[ "$line" == :* ]]; then
                break
            fi
            if [[ "$line" == *$'\t'* ]]; then
                local completion="${line%%%%$'\t'*}"
                local description="${line#*$'\t'}"
                completions_with_descriptions+=("$completion:$description")
            else
                completions+=("$line")
            fi
        done

        if [ ${#completions_with_descriptions[@]} -gt 0 ]; then
            _describe -t completions 'gh %s' completions_with_descriptions
        fi

        if [ ${#completions[@]} -gt 0 ]; then
            compadd -a completions
        fi
    }
else
    echo "Warning: gh completion not loaded. Please install gh completion first." >&2
    echo "Run: gh completion -s zsh > \"\${fpath[1]}/_gh\" && compinit" >&2
fi
`, execName, extName, extName, funcName, funcName, extName, extName)
}

func generateFishPatch(execName, extName string) string {
	return fmt.Sprintf(`# %s extension completion patch for fish
# Source this file after gh completion to enable 'gh %s' completion

# Define completion function for gh %s
function __%s_complete
    set -l tokens (commandline -opc)
    set -l current (commandline -ct)

    # Extract arguments after 'gh %s'
    set -l args
    if test (count $tokens) -ge 3
        set args $tokens[3..-1]
    end

    # Call the extension's __complete command
    gh %s __complete $args "$current" 2>/dev/null | string match -v ':*'
end

# Check if we're completing gh %s subcommands
function __%s_is_completing
    set -l tokens (commandline -opc)
    set -l current_token (commandline -ct)

    # Must have at least 'gh %s' on the command line
    if test (count $tokens) -lt 2
        return 1
    end

    # Second token must be '%s'
    if test "$tokens[2]" != "%s"
        return 1
    end

    return 0
end

# Add '%s' as a subcommand option for 'gh'
complete -c gh -n "__fish_use_subcommand" -a %s -d "Manage teams (extension)"

# Provide completions for 'gh %s' subcommands
# -f: disable file completion
# -n: only when __team-kit_is_completing condition is true
complete -c gh -f -n "__%s_is_completing" -a "(__%s_complete)"
`, execName, extName, extName, extName, extName, extName, extName, extName, extName, extName, extName, extName, extName, extName, extName, extName)
}

func generatePowerShellPatch(execName, extName string) string {
	return fmt.Sprintf(`# %s extension completion patch for PowerShell
# Source this file after gh completion to enable 'gh %s' completion

# Save original gh completer if it exists
$__original_gh_completer = $null
if (Get-Variable -Name __ghCompleterBlock -ErrorAction SilentlyContinue) {
    $__original_gh_completer = ${__ghCompleterBlock}
}

# Override gh completion to handle %s
Register-ArgumentCompleter -CommandName gh -ScriptBlock {
    param($wordToComplete, $commandAst, $cursorPosition)

    # Get the command line as tokens
    $tokens = $commandAst.CommandElements
    $words = @()
    foreach ($token in $tokens) {
        $words += $token.ToString()
    }

    # Check if this is a %s completion
    if ($words.Count -gt 1 -and $words[1] -eq '%s') {
        # Build args for __complete
        $completeArgs = @()
        if ($words.Count -gt 2) {
            $completeArgs = $words[2..($words.Count - 1)]
        }
        if (-not $wordToComplete) {
            $completeArgs += ""
        }

        # Call gh %s __complete
        $output = & gh %s __complete @completeArgs 2>$null

        # Parse output into completion results
        $results = @()
        foreach ($line in $output) {
            if ($line -match '^([^\t]+)\t(.+)$') {
                $results += [System.Management.Automation.CompletionResult]::new($matches[1], $matches[1], 'ParameterValue', $matches[2])
            } elseif ($line -notmatch '^:') {
                $results += [System.Management.Automation.CompletionResult]::new($line, $line, 'ParameterValue', $line)
            }
        }
        return $results
    } elseif ($__original_gh_completer) {
        # Use original gh completer
        return & $__original_gh_completer $wordToComplete $commandAst $cursorPosition
    } else {
        # Fallback: return nothing to let default completion work
        return @()
    }
}
`, execName, extName, extName, extName, extName, extName, extName)
}
