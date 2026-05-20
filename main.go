package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"splito/parser"
	"splito/report"
)

const usage = `splito — Orienteering Results Tool

Usage:
  splito.exe [options] <input.xml>

Options:
  -class    string   Class name to process (e.g. "M 19 A").
                     Omit to list classes. Use with -all to generate all classes.
  -all               Generate a single HTML with all classes (ignores -class).
  -mode     string   Output mode: "splits", "timelost", or "all" (default: "all")
  -out      string   Output HTML path (default: auto-named next to input file)
  -help              Show this help

Examples:
  splito.exe results.xml                         (list classes)
  splito.exe -class "M 19 A" results.xml         (one class, all sections)
  splito.exe -all results.xml                    (all classes in one file)
  splito.exe -all -mode splits results.xml       (all classes, splits only)
`

func main() {
	classFlag := flag.String("class", "", "Class name to process")
	allFlag := flag.Bool("all", false, "Generate report for all classes in one file")
	modeFlag := flag.String("mode", "all", `Output mode: "splits", "timelost", or "all"`)
	outFlag := flag.String("out", "", "Output HTML file path")
	helpFlag := flag.Bool("help", false, "Show help")

	flag.Usage = func() { fmt.Fprint(os.Stderr, usage) }
	flag.Parse()

	if *helpFlag {
		fmt.Print(usage)
		os.Exit(0)
	}

	args := flag.Args()
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Error: no input XML file specified.")
		fmt.Fprint(os.Stderr, usage)
		os.Exit(1)
	}
	inputFile := args[0]

	// Validate mode
	mode := strings.ToLower(*modeFlag)
	if mode != "splits" && mode != "timelost" && mode != "all" {
		fmt.Fprintf(os.Stderr, "Error: unknown mode %q. Use \"splits\", \"timelost\", or \"all\".\n", *modeFlag)
		os.Exit(1)
	}

	// Parse the XML
	fmt.Printf("Parsing %s ...\n", inputFile)
	classes, classOrder, err := parser.ParseFile(inputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Determine event name from first class
	eventName := ""
	if len(classOrder) > 0 {
		eventName = classes[classOrder[0]].EventName
	}

	base := strings.TrimSuffix(inputFile, filepath.Ext(inputFile))

	// ---- ALL CLASSES mode ----
	if *allFlag {
		outPath := *outFlag
		if outPath == "" {
			outPath = fmt.Sprintf("%s_all_%s.html", base, mode)
		}
		fmt.Printf("Generating report for all %d classes (mode: %s) ...\n", len(classOrder), mode)
		err = report.GenerateAllClassesHTML(classes, classOrder, eventName, outPath, mode)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error generating report: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Done! Output saved to: %s\n", outPath)
		os.Exit(0)
	}

	// ---- No class specified: list ----
	if *classFlag == "" {
		fmt.Printf("\nAvailable classes in %s:\n\n", filepath.Base(inputFile))
		for _, name := range classOrder {
			cd := classes[name]
			okCount := 0
			dnfCount := 0
			for _, c := range cd.Competitors {
				if c.Status == "OK" {
					okCount++
				} else if c.Status == "DNF" || c.Status == "MP" {
					dnfCount++
				}
			}
			fmt.Printf("  %-20s  %d controls  %d finishers  %d DNF\n",
				name, len(cd.Controls), okCount, dnfCount)
		}
		fmt.Println()
		fmt.Println("Use -class \"<name>\" or -all to generate a report.")
		os.Exit(0)
	}

	// ---- Single class mode ----
	cd, ok := classes[*classFlag]
	if !ok {
		fmt.Fprintf(os.Stderr, "Error: class %q not found.\n\nAvailable classes:\n", *classFlag)
		for _, name := range classOrder {
			fmt.Fprintf(os.Stderr, "  %s\n", name)
		}
		os.Exit(1)
	}

	outPath := *outFlag
	if outPath == "" {
		safeName := strings.Map(func(r rune) rune {
			if r == ' ' {
				return '_'
			}
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '-' {
				return r
			}
			return '_'
		}, *classFlag)
		outPath = fmt.Sprintf("%s_%s_%s.html", base, safeName, mode)
	}

	fmt.Printf("Generating report for class %q (mode: %s) ...\n", *classFlag, mode)
	err = report.GenerateHTML(cd, outPath, mode)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating report: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Done! Output saved to: %s\n", outPath)
}
