/*
GuardDuty Findings Exporter

Author: Binh Phan (Binhphan@me.com)

This program is a web application that allows users to export AWS GuardDuty findings
from multiple US regions into a CSV file. It provides a simple web interface for
selecting regions and initiating the export process.

Key features:
- Web-based interface for easy interaction
- Dynamically fetches and displays available US AWS regions
- Allows selection of multiple regions for export
- Exports GuardDuty findings to a CSV file
- Provides real-time progress updates during the export process

Usage:
1. Run the program: go run main.go
2. Open a web browser and navigate to http://localhost:8080
3. Select desired US regions and click "Export Findings"
4. Wait for the export to complete and download the CSV file

Note: Ensure AWS credentials are properly configured before running the program.
*/

package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/guardduty"
	"github.com/aws/aws-sdk-go-v2/service/guardduty/types"
)

// Global AWS configuration
var cfg aws.Config

func main() {
	// Load the AWS SDK configuration
	var err error
	cfg, err = config.LoadDefaultConfig(context.TODO())
	if err != nil {
		fmt.Printf("Unable to load SDK config, %v\n", err)
		return
	}

	// Set up HTTP routes
	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/api/regions", handleRegions)
	http.HandleFunc("/api/export", handleExport)

	// Start the HTTP server
	fmt.Println("Server is running on http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}

// handleIndex serves the main HTML page
func handleIndex(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}

// handleRegions returns a list of US regions as JSON
func handleRegions(w http.ResponseWriter, r *http.Request) {
	regions, err := getUSRegions(cfg)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(regions)
}

// handleExport generates a CSV file with GuardDuty findings from selected regions
func handleExport(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Export process started")

	regions := r.URL.Query()["regions"]
	if len(regions) == 0 {
		http.Error(w, "No regions specified", http.StatusBadRequest)
		return
	}

	fmt.Printf("Selected regions: %v\n", regions)

	filename := fmt.Sprintf("guardduty_findings_%s.csv", time.Now().Format("20060102_150405"))
	file, err := os.Create(filename)
	if err != nil {
		fmt.Printf("Error creating file: %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	header := []string{"Region", "FindingId", "Title", "Description", "Severity", "CreatedAt", "UpdatedAt"}
	if err := writer.Write(header); err != nil {
		fmt.Printf("Error writing CSV header: %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	totalFindings := 0
	for _, region := range regions {
		fmt.Printf("Starting export for region: %s\n", region)
		findings, err := getGuardDutyFindings(cfg, region)
		if err != nil {
			fmt.Printf("Error getting findings for region %s: %v\n", region, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Printf("Writing %d findings for region %s\n", len(findings), region)
		for _, finding := range findings {
			row := []string{
				region,
				*finding.Id,
				*finding.Title,
				*finding.Description,
				fmt.Sprintf("%.1f", *finding.Severity),
				*finding.CreatedAt,
				*finding.UpdatedAt,
			}
			if err := writer.Write(row); err != nil {
				fmt.Printf("Error writing finding to CSV: %v\n", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		totalFindings += len(findings)
		fmt.Printf("Completed region %s. Total findings so far: %d\n", region, totalFindings)
	}

	fmt.Printf("Export completed. Total findings across all regions: %d. File: %s\n", totalFindings, filename)
	w.Write([]byte(filename))
}

// getUSRegions returns a list of US AWS regions
func getUSRegions(cfg aws.Config) ([]string, error) {
	client := ec2.NewFromConfig(cfg)
	resp, err := client.DescribeRegions(context.TODO(), &ec2.DescribeRegionsInput{})
	if err != nil {
		return nil, err
	}

	var regions []string
	for _, region := range resp.Regions {
		// Filter for US regions only
		if aws.ToString(region.RegionName)[:2] == "us" {
			regions = append(regions, aws.ToString(region.RegionName))
		}
	}
	return regions, nil
}

// getGuardDutyFindings fetches GuardDuty findings for a specific region
func getGuardDutyFindings(cfg aws.Config, region string) ([]types.Finding, error) {
	fmt.Printf("Fetching GuardDuty findings for region: %s\n", region)

	cfg.Region = region
	client := guardduty.NewFromConfig(cfg)

	detectors, err := client.ListDetectors(context.TODO(), &guardduty.ListDetectorsInput{})
	if err != nil {
		return nil, fmt.Errorf("error listing detectors in region %s: %v", region, err)
	}

	fmt.Printf("Found %d detectors in region %s\n", len(detectors.DetectorIds), region)

	var allFindings []types.Finding
	for _, detectorID := range detectors.DetectorIds {
		fmt.Printf("Processing detector: %s\n", detectorID)
		paginator := guardduty.NewListFindingsPaginator(client, &guardduty.ListFindingsInput{
			DetectorId: aws.String(detectorID),
		})

		pageCount := 0
		for paginator.HasMorePages() {
			pageCount++
			fmt.Printf("Processing page %d for detector %s\n", pageCount, detectorID)

			output, err := paginator.NextPage(context.TODO())
			if err != nil {
				return nil, fmt.Errorf("error listing findings for detector %s: %v", detectorID, err)
			}

			if len(output.FindingIds) > 0 {
				fmt.Printf("Found %d findings on page %d for detector %s\n", len(output.FindingIds), pageCount, detectorID)
				getFindingsInput := &guardduty.GetFindingsInput{
					DetectorId: aws.String(detectorID),
					FindingIds: output.FindingIds,
				}
				getFindingsOutput, err := client.GetFindings(context.TODO(), getFindingsInput)
				if err != nil {
					return nil, fmt.Errorf("error getting detailed findings for detector %s: %v", detectorID, err)
				}
				allFindings = append(allFindings, getFindingsOutput.Findings...)
			} else {
				fmt.Printf("No findings on page %d for detector %s\n", pageCount, detectorID)
			}
		}
		fmt.Printf("Finished processing detector %s. Total pages: %d\n", detectorID, pageCount)
	}

	fmt.Printf("Total findings for region %s: %d\n", region, len(allFindings))
	return allFindings, nil
}
