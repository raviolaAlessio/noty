package chart

import (
	"github.com/spf13/cobra"
)

func init() {
	ChartCmd.AddCommand(ChartSprints)
}

var ChartCmd = &cobra.Command{
	Use:   "chart",
	Short: "generate useful charts",
}
