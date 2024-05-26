# go-tools.nvim
A Neovim plugin to generate common things in Go. Invoked via a single keybinding based on context.
If the context is ambiguous, a selector is displayed.

## Supported Operations
- [x] Generate constructor
- [ ] Generate `if err != nil { ... }`

## Installation

### lazy.nvim

```lua
{ 'cszczepaniak/go-tools.nvim' }
```
