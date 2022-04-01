package cmd

import (
	"bufio"
	"log"
	"os"

	"github.com/djboris9/rea/pkg/odf"
	"github.com/djboris9/rea/pkg/template"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(templateCmd)
}

var templateCmd = &cobra.Command{
	Use:   "template",
	Short: "Process a template document to generate a filled out document",
	//Long:  `TODO`,
	Run: templateCmdRun,
}

func templateCmdRun(cmd *cobra.Command, args []string) {
	// Get flag variables
	tmplFile, err := cmd.Flags().GetString("template")
	if err != nil {
		log.Fatalf("reading template flag: %s", err)
	}

	//inputFile, err := cmd.Flags().GetString("input")
	//if err != nil {
	//	log.Fatalf("reading input flag: %w", err)
	//}

	outputFile, err := cmd.Flags().GetString("output")
	if err != nil {
		log.Fatalf("reading output flag: %s", err)
	}

	// Open files
	output, err := os.Create(outputFile)
	if err != nil {
		log.Fatalf("creating output file %s: %s", outputFile, err)
	}
	outputBuf := bufio.NewWriter(output)

	ott, err := odf.NewFromFile(tmplFile)
	if err != nil {
		log.Fatalf("loading template file %s: %s", tmplFile, err)
	}

	// Run rendering
	err = template.TemplateODT(ott, nil, outputBuf) // TODO: data
	if err != nil {
		log.Fatalf("executing templating: %s", err)
	}

	// Finish
	err = outputBuf.Flush()
	if err != nil {
		log.Fatalf("flushing output buffer: %s", err)
	}

	err = output.Close()
	if err != nil {
		log.Fatalf("closing output file: %s", err)
	}
}

func init() {
	templateCmd.Flags().StringP("template", "t", "template.ott", "template document")
	templateCmd.Flags().StringP("input", "i", "data.yaml", "data file")
	templateCmd.Flags().StringP("output", "o", "document.odt", "output document")
}
