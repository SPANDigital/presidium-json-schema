package cmd

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/SPANDigital/presidium-json-schema/pkg/markdown"
	"github.com/spf13/cobra"
)

var config markdown.Config

func init() {
	flags := convert.Flags()
	flags.StringVarP(&config.Destination, "destination", "d", ".", "the output directory")
	flags.StringVarP(&config.Extension, "extension", "e", "*.schema.json", "the schema extension")
	flags.BoolVarP(&config.Recursive, "walk", "w", false, "walk through sub-directories")
	flags.BoolVarP(&config.Ordered, "ordered", "o", false, "preserve the schema order (defaults to alphabetical)")
	flags.BoolVarP(&config.OrderedFilePath, "orderedfilepath", "p", false, "preserve the schema order (defaults to alphabetical) by appending a digit to the filename prefix")
	flags.BoolVarP(&config.Clean, "clean", "c", false, "removes the output directory before generating output files")
	rootCmd.AddCommand(convert)
}

var convert = &cobra.Command{
	Use:   "convert [path]",
	Short: "convert [path]",
	Args:  validatePaths(),
	Run: func(cmd *cobra.Command, args []string) {
		c := markdown.NewConverter(config)
		if err := c.Convert(args[0]); err != nil {
			log.Fatal(err)
		}
		return
	},
}

func validatePaths() cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return errors.New("requires at least 1 folder path")
		}

		for _, path := range args {
			if _, err := os.Stat(path); os.IsNotExist(err) {
				return fmt.Errorf(`provided path "%s" does not exist`, path)
			}
		}
		return nil
	}
}
