package cmd

import (
	"fmt"
	"strings"

	"github.com/piyushgupta/polymarket-cli/internal/output"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// commandInfo describes a CLI command for agent discovery.
type commandInfo struct {
	Name         string       `json:"name"`
	Description  string       `json:"description"`
	Flags        []flagInfo   `json:"flags,omitempty"`
	RequiresAuth bool         `json:"requires_auth"`
	Subcommands  []string     `json:"subcommands,omitempty"`
}

// flagInfo describes a single flag on a command.
type flagInfo struct {
	Name     string `json:"name"`
	Short    string `json:"short,omitempty"`
	Type     string `json:"type"`
	Default  string `json:"default,omitempty"`
	Required bool   `json:"required,omitempty"`
	Usage    string `json:"usage"`
}

// authCommands is the set of command paths that require authentication.
var authCommands = map[string]bool{
	// wallet
	"wallet address": true,
	"wallet show":    true,
	"wallet reset":   true,
	// clob trading
	"clob create-order":  true,
	"clob cancel":        true,
	"clob cancel-orders": true,
	"clob cancel-market": true,
	"clob cancel-all":    true,
	"clob order":         true,
	"clob orders":        true,
	"clob trades":        true,
	"clob balance":       true,
	"clob update-balance": true,
	// clob rewards
	"clob rewards":            true,
	"clob earnings":           true,
	"clob earnings-total":     true,
	"clob reward-percentages": true,
	"clob market-reward":      true,
	"clob order-scoring":      true,
	"clob orders-scoring":     true,
	// clob account
	"clob api-keys":             true,
	"clob create-api-key":       true,
	"clob delete-api-key":       true,
	"clob notifications":        true,
	"clob delete-notifications": true,
	// on-chain
	"approve check":       true,
	"approve set":         true,
	"ctf split":           true,
	"ctf merge":           true,
	"ctf redeem":          true,
	"ctf redeem-neg-risk": true,
}

var commandsCmd = &cobra.Command{
	Use:   "commands",
	Short: "Discover available CLI commands",
	Long:  "List and describe all available commands. Designed for agent/programmatic discovery.",
}

var commandsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available commands",
	Example: `  # List all commands
  polymarket commands list

  # Agent: get command list as JSON
  polymarket commands list -o json -q`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var commands []commandInfo
		walkCommands(rootCmd, "", &commands)

		if getOutputFormat() == "json" {
			return output.PrintJSON(commands)
		}

		// Table mode
		for _, c := range commands {
			authLabel := ""
			if c.RequiresAuth {
				authLabel = " [auth]"
			}
			fmt.Printf("  %-35s %s%s\n", c.Name, c.Description, authLabel)
		}
		return nil
	},
}

var commandsDescribeCmd = &cobra.Command{
	Use:   "describe <command-path>",
	Short: "Describe a specific command",
	Long:  "Show detailed information about a command including all flags.\nUse space-separated path, e.g. 'markets list'.",
	Example: `  # Describe a command
  polymarket commands describe "markets list"

  # Agent: get command metadata as JSON
  polymarket commands describe "clob create-order" -o json -q`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		target := findCommand(rootCmd, args[0])
		if target == nil {
			return output.ErrNotFound(fmt.Sprintf("command %q", args[0]))
		}

		info := describeCommand(target, args[0])

		if getOutputFormat() == "json" {
			return output.PrintJSON(info)
		}

		// Table mode
		fmt.Printf("Command: %s\n", info.Name)
		fmt.Printf("Description: %s\n", info.Description)
		fmt.Printf("Requires Auth: %v\n", info.RequiresAuth)
		if len(info.Subcommands) > 0 {
			fmt.Printf("Subcommands: %s\n", strings.Join(info.Subcommands, ", "))
		}
		if len(info.Flags) > 0 {
			fmt.Println("\nFlags:")
			for _, f := range info.Flags {
				short := ""
				if f.Short != "" {
					short = fmt.Sprintf(" (-%s)", f.Short)
				}
				def := ""
				if f.Default != "" {
					def = fmt.Sprintf(" [default: %s]", f.Default)
				}
				req := ""
			if f.Required {
				req = " [required]"
			}
			fmt.Printf("  --%s%s  %s  %s%s%s\n", f.Name, short, f.Type, f.Usage, def, req)
	}
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(commandsCmd)
	commandsCmd.AddCommand(commandsListCmd)
	commandsCmd.AddCommand(commandsDescribeCmd)
}

// walkCommands recursively collects leaf commands (those with RunE/Run).
func walkCommands(cmd *cobra.Command, prefix string, result *[]commandInfo) {
	for _, child := range cmd.Commands() {
		if child.Hidden || child.Name() == "help" || child.Name() == "completion" {
			continue
		}

		path := child.Name()
		if prefix != "" {
			path = prefix + " " + child.Name()
		}

		if child.HasSubCommands() {
			// Collect as a group node with subcommands listed
			walkCommands(child, path, result)
		}

		// Include if it has a Run/RunE (it's executable)
		if child.RunE != nil || child.Run != nil {
			*result = append(*result, commandInfo{
				Name:         path,
				Description:  child.Short,
				RequiresAuth: commandRequiresAuth(path),
			})
		}
	}
}

// commandRequiresAuth checks if a command path requires authentication.
func commandRequiresAuth(path string) bool {
	return authCommands[path]
}

// findCommand navigates the cobra tree by a space-separated path.
func findCommand(root *cobra.Command, path string) *cobra.Command {
	parts := strings.Fields(path)
	current := root
	for _, part := range parts {
		found := false
		for _, child := range current.Commands() {
			if child.Name() == part {
				current = child
				found = true
				break
			}
		}
		if !found {
			return nil
		}
	}
	if current == root {
		return nil
	}
	return current
}

// describeCommand builds a full commandInfo including flags and subcommands.
func describeCommand(cmd *cobra.Command, path string) commandInfo {
	info := commandInfo{
		Name:         path,
		Description:  cmd.Short,
		RequiresAuth: commandRequiresAuth(path),
	}

	// Collect subcommands
	for _, child := range cmd.Commands() {
		if child.Hidden || child.Name() == "help" || child.Name() == "completion" {
			continue
		}
		info.Subcommands = append(info.Subcommands, child.Name())
	}

	// Collect flags (local only, not inherited persistent flags)
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		if f.Hidden {
			return
		}
		fi := flagInfo{
			Name:  f.Name,
			Type:  f.Value.Type(),
			Usage: f.Usage,
		}
		if f.Shorthand != "" {
			fi.Short = f.Shorthand
		}
		if f.DefValue != "" && f.DefValue != "false" && f.DefValue != "0" {
			fi.Default = f.DefValue
		}
		if _, ok := f.Annotations[cobra.BashCompOneRequiredFlag]; ok {
			fi.Required = true
		}
		info.Flags = append(info.Flags, fi)
	})

	return info
}
