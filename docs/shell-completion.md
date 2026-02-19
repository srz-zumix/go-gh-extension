# Shell Completion for gh Extensions

This guide explains how to enable shell completion for gh CLI extensions that use the `go-gh-extension/pkg/completion` package.

## Prerequisites

Before setting up extension completion, you must first configure gh CLI completion for your shell. The extension completion patch extends gh's existing completion system.

See the [gh completion documentation](https://cli.github.com/manual/gh_completion) for detailed setup instructions.

**Quick setup:**

- **Bash**: `gh completion -s bash >> ~/.bashrc` (or `~/.bash_profile` on macOS)
- **Zsh**: `gh completion -s zsh > "${fpath[1]}/_gh"` then run `compinit`
- **Fish**: `gh completion -s fish > ~/.config/fish/completions/gh.fish`
- **PowerShell**: Add `Invoke-Expression -Command $(gh completion -s powershell | Out-String)` to your profile

> **Note**: For PowerShell, you must load gh completion in your session before loading extension completion.

## Setup

While gh CLI doesn't natively support extension completion, we provide a workaround that patches gh's completion to route requests directly to the extension.

Replace `<extension>` in the examples below with the actual extension name (e.g., `team-kit`, `label-kit`).

### Apply the completion patch

The completion command automatically detects when called as `gh <extension>` and generates the appropriate patch.

**Bash:**

First, ensure that you install bash-completion using your package manager.

After, add this to your ~/.bash_profile:

```sh
eval "$(gh <extension> completion -s bash)"
```

**Zsh:**

Add this to your ~/.zshrc:

```sh
eval "$(gh <extension> completion -s zsh)"
```

Or, if you prefer to generate a completion file:

```sh
# If _gh already exists in your fpath, append to it
gh <extension> completion -s zsh >> "${fpath[1]}/_gh"

# Then reload
compinit
```

> **Note**: If you use the file method and already have gh completion, use `>>` (append) instead of `>` (overwrite) to preserve existing gh completions.

**Fish:**

Fish completion requires patching the gh completion file:

```sh
# Backup original gh completion
cp ~/.config/fish/completions/gh.fish ~/.config/fish/completions/gh.fish.backup

# Append extension completion to gh.fish
gh <extension> completion -s fish >> ~/.config/fish/completions/gh.fish
```

Then open a new terminal or reload completions:

```sh
# Open a new terminal, or manually reload in current session:
source ~/.config/fish/completions/gh.fish
```

> **Note**: Fish loads `completions/gh.fish` for all `gh` commands. Since gh's completion system doesn't support extensions, we append our completion rules to the existing gh.fish file. If you regenerate gh completion (`gh completion -s fish`), you'll need to re-append the extension completion.

**PowerShell:**

Open your profile script with:

```powershell
mkdir -Path (Split-Path -Parent $profile) -ErrorAction SilentlyContinue
notepad $profile
```

Add both gh and extension completion to your profile (in this order):

```powershell
# Load gh completion first (required)
Invoke-Expression -Command $(gh completion -s powershell | Out-String)

# Then load extension completion patch
Invoke-Expression -Command $(gh <extension> completion -s powershell | Out-String)
```

> **Important**: PowerShell requires gh completion to be loaded before extension completion.

### Multiple extensions

Multiple extensions using this package can be loaded simultaneously. Each extension registers its own completion patch with a unique identifier, forming a chain:

```sh
# Load multiple extension completions (order doesn't matter)
eval "$(gh team-kit completion -s bash)"
eval "$(gh label-kit completion -s bash)"
```

Each patch is idempotent — loading the same extension twice has no effect.

### Test the completion

```sh
gh <extension> <TAB>         # Shows available subcommands
gh <extension> <subcommand> <TAB>  # Shows subcommand options
```

## How it works

The patch script intercepts completion requests in gh's completion function and routes them to the extension when it detects the extension name:

```bash
# Instead of: gh __complete <extension> <subcommand> ""  (returns nothing)
# It calls:   gh <extension> __complete <subcommand> ""  (works correctly)
```

### Technical Details

The gh CLI completion script calls `gh __complete <args>` for all completions. However, gh does not route these calls to extensions:

```sh
# This returns nothing (gh CLI limitation)
gh __complete <extension> <subcommand> ""
# Output: :0

# This works (direct call to extension)
gh <extension> __complete <subcommand> ""
# Output: subcommand completions :4
```

Our workaround patches gh's completion function to detect the extension name and route the request to `gh <extension> __complete` instead of `gh __complete`. Each extension saves the previous completion function with a unique name (e.g., `__gh_get_completion_results_before_<func_name>`), enabling multiple extensions to coexist by forming a delegation chain.
