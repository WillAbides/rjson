package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/willabides/benchdiff/pkg/benchstatter"
	"golang.org/x/perf/benchstat"
)

func runComp(w io.Writer, mainFile, compFile string) error {
	stat := &benchstatter.Benchstat{
		OutputFormatter: benchstatter.MarkdownFormatter(&benchstatter.MarkdownFormatterOptions{
			benchstatter.CSVFormatterOptions{
				NoRange: true,
			},
		}),
	}
	res, err := stat.Run(compFile, mainFile)
	if err != nil {
		return err
	}
	var tables []*benchstat.Table
	for _, table := range res.Tables() {
		if table.Metric != "speed" && table.Metric != "allocs/op" {
			tables = append(tables, table)
		}
	}
	for _, table := range tables {
		table.Metric = ""
		for _, row := range table.Rows {
			row.Note = ""
		}
	}
	var buf bytes.Buffer
	err = stat.OutputTables(&buf, tables)
	if err != nil {
		return err
	}
	out := buf.String()
	out = strings.ReplaceAll(out, `| old  `, fmt.Sprintf(`| %s `, filepath.Base(compFile)))
	out = strings.ReplaceAll(out, `| new  `, fmt.Sprintf(`| %s `, filepath.Base(mainFile)))
	_, err = w.Write([]byte(out + "\n"))
	return err
}

func runComps(w io.Writer, mainFile string, comps []string) error {
	for _, comp := range comps {
		_, err := w.Write([]byte(fmt.Sprintf("### %s\n\n", filepath.Base(comp))))
		if err != nil {
			return err
		}
		err = runComp(w, mainFile, comp)
		if err != nil {
			return err
		}
	}
	return nil
}

func main() {
	if len(os.Args) < 3 {
		log.Fatal("at least two files are needed")
	}

	mainFile := os.Args[1]
	comps := os.Args[2:]

	err := runComps(os.Stdout, mainFile, comps)
	if err != nil {
		log.Fatal(err)
	}
}
