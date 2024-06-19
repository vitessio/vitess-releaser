## Local Development

This document describes how to set up the environment for local development of the Vitess Releaser tool. Note that 
the tool UI is intended to be self-documenting and authoritative. So if there are any discrepancies between this
document and the tool UI, the tool UI should be considered the source of truth.

For testing local changes you need to create two subdirectories:

1. Vitess fork
   Create a fork of Vitess in your personal GitHub account and clone it to the `vitess` subdirectory. When you don't use
   the `--live` flag, the tool will use this fork to create the release branches and PRs.

2. Vitess Operator
   Create a fork of Vitess Operator in your personal GitHub account and clone it to the `vitess-operator` subdirectory.
   When you don't use the `--live` flag, the tool will use this fork to create the release branches and PRs.

### Testing

`vitess-releaser --date="2024-06-19" --rc=3 --release=20`

You need to chose a valid release: here there should be a tag `v20.0.0-RC3`.

### Vitess fork

Just creating a fork is not enough. The tool expects the following to be setup on your fork, which you will need to do
manually:
### Labels
Create the following labels on your fork: `Component: General` and `Type: Release`. This is required because the 
Releaser tool uses these labels to create the Release issue. The issue URL is shown in the UI that comes up. To test 
with the same 

### Miscellaneous Notes
* We use [bubbletea](https://github.com/charmbracelet/bubbletea) for the UI 
with [lipgloss](https://github.com/charmbracelet/lipgloss) for styling.