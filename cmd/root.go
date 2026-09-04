package cmd

import (
	"fmt"
	"os"
	"strings"

	"closed-sessions/internal/closedsessions"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "closed-sessions [date]",
	Short: "List restore-session files for a given date or relative duration",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := closedsessions.LoadConfig()
		if err != nil {
			return err
		}

		if flagValue, err := cmd.Flags().GetString("base-search-path"); err == nil && strings.TrimSpace(flagValue) != "" {
			cfg.BaseSearchPath = flagValue
			viper.Set("base_search_path", flagValue)
		}

		dateArg := ""
		if len(args) > 0 {
			dateArg = args[0]
		}

		results, err := closedsessions.Search(cfg.BaseSearchPath, dateArg)
		if err != nil {
			return err
		}

		if len(results) == 0 {
			fmt.Printf("No matching session files found under %s\n", cfg.BaseSearchPath)
			return nil
		}

		output := closedsessions.RenderTable(results)
		fmt.Print(output)
		return nil
	},
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage the closed-sessions config",
}

var configSetCmd = &cobra.Command{
	Use:   "set [base-search-path]",
	Short: "Set the base directory to search for files",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		value := strings.TrimSpace(args[0])
		if value == "" {
			return fmt.Errorf("base-search-path cannot be empty")
		}
		if err := closedsessions.SaveConfig(value); err != nil {
			return err
		}
		fmt.Printf("Saved base search path: %s\n", value)
		return nil
	},
}

var configGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Print the configured base search path",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := closedsessions.LoadConfig()
		if err != nil {
			return err
		}
		fmt.Println(cfg.BaseSearchPath)
		return nil
	},
}

func init() {
	viper.SetDefault("base_search_path", closedsessions.DefaultBaseSearchPath())
	rootCmd.PersistentFlags().String("base-search-path", "", "Base path to search for closed session files")
	_ = viper.BindPFlag("base_search_path", rootCmd.PersistentFlags().Lookup("base-search-path"))
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configGetCmd)
	rootCmd.AddCommand(configCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
