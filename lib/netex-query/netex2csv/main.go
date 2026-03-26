// SPDX-FileCopyrightText: 2024 NOI Techpark <digital@noi.bz.it>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/noi-techpark/opendatahub-public-transport/lib/netex-query/netex"
	_ "github.com/noi-techpark/opendatahub-public-transport/lib/netex-query/netex/profile" // register profiles via init()
)

func main() {
	inputPath := flag.String("input", "", "Path to NeTEx XML file (required)")
	outputDir := flag.String("output", "output", "Output directory for CSV files")
	profileName := flag.String("profile", "epip", "Profile name (e.g., epip, it-l2)")
	flag.Parse()

	if *inputPath == "" {
		fmt.Fprintln(os.Stderr, "error: -input is required")
		flag.Usage()
		os.Exit(1)
	}

	// Get profile
	prof, err := netex.GetProfile(*profileName)
	if err != nil {
		log.Fatalf("profile: %v", err)
	}
	fmt.Printf("Using profile: %s\n", prof.Name())

	// Create CSV output
	csvOut, err := netex.NewCSVOutput(*outputDir, prof.Tables())
	if err != nil {
		log.Fatalf("csv output: %v", err)
	}
	defer csvOut.CloseAll()

	// Open input file
	f, err := os.Open(*inputPath)
	if err != nil {
		log.Fatalf("open input: %v", err)
	}
	defer f.Close()

	// Parse
	start := time.Now()
	fmt.Printf("Parsing %s...\n", *inputPath)

	if err := netex.Parse(f, prof, csvOut.Handle); err != nil {
		log.Fatalf("parse: %v", err)
	}

	elapsed := time.Since(start)
	fmt.Printf("\nDone in %v\n\n", elapsed.Round(time.Millisecond))
	fmt.Println("Output:")
	csvOut.PrintSummary()
}
