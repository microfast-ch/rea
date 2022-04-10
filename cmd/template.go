package cmd

import (
	"bufio"
	"log"
	"os"

	"github.com/djboris9/rea/internal/factory"
	"github.com/djboris9/rea/internal/writer"
	"github.com/djboris9/rea/pkg/bundle"
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

	bundleFile, err := cmd.Flags().GetString("bundle")
	if err != nil {
		log.Fatalf("reading bundle flag: %s", err)
	}

	debug, err := cmd.Flags().GetBool("debug")
	if err != nil {
		log.Fatalf("reading debug flag: %s", err)
	}

	// inputFile, err := cmd.Flags().GetString("input")
	// if err != nil {
	// 	log.Fatalf("reading input flag: %w", err)
	// }

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

	docTemplate, err := factory.NewFromFile(tmplFile)
	if err != nil {
		log.Fatalf("error loading template file %s: %s", tmplFile, err)
	}

	// Create bundle writer
	var bundleW *bundle.Writer

	if bundleFile != "" {
		bundleFD, err := os.Create(bundleFile)
		if err != nil {
			log.Fatalf("creating bundle file %s: %s", bundleFile, err)
		}

		bundleW = bundle.New(bundleFD, debug)
	}

	// Run rendering and first write bundle before throwing error
	tpd, err := writer.Write(docTemplate, nil, outputBuf) // TODO: data
	if err != nil {
		log.Fatalf("executing templating: %s", err)
	}

	if bundleW != nil && tpd != nil {
		bundleW.AddTemplateMimeType(tpd.TemplateMimeType)
		bundleW.AddLuaProg(tpd.TemplateLuaProg)
		bundleW.AddLuaNodeList(tpd.TemplateLuaNodeList)
		bundleW.AddTemplateXMLTree(tpd.TemplateXMLTree)
		bundleW.AddLuaNodePathStr(tpd.LuaNodePathStr)
		bundleW.AddContentXML(tpd.ContentXML)

		if errB := bundleW.Close(); errB != nil {
			log.Printf("closing bundle writer: %s", errB)
		}
	}

	// Finish
	err = outputBuf.Flush()
	if err != nil {
		log.Fatalf("error flushing output buffer: %s", err)
	}

	err = output.Close()
	if err != nil {
		log.Fatalf("error closing output file: %s", err)
	}
}

func init() {
	templateCmd.Flags().StringP("template", "t", "template.ott", "template document")
	templateCmd.Flags().StringP("input", "i", "data.yaml", "data file")
	templateCmd.Flags().StringP("output", "o", "document.odt", "output document")
	templateCmd.Flags().StringP("bundle", "b", "", "tar file to which the job bundle should be written")
	templateCmd.Flags().BoolP("debug", "d", false, "write debug information to job bundle")
}
