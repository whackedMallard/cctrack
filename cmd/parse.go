package cmd

import (
	"fmt"

	"github.com/kylecalbert/cctrack/internal/config"
	"github.com/kylecalbert/cctrack/internal/parser"
	"github.com/kylecalbert/cctrack/internal/store"
	"github.com/spf13/cobra"
)

var parseCmd = &cobra.Command{
	Use:   "parse",
	Short: "Manually trigger log parsing",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		s, err := store.Open(cfg.DBPath)
		if err != nil {
			return fmt.Errorf("opening store: %w", err)
		}
		defer s.Close()

		p := parser.New(s)
		files, sessions, err := p.ParseAll(cfg.LogDir)
		if err != nil {
			return fmt.Errorf("parsing: %w", err)
		}

		fmt.Printf("Parsed %d files, %d sessions updated\n", files, sessions)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(parseCmd)
}
