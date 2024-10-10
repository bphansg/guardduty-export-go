# GuardDuty Findings Exporter

## Description
GuardDuty Findings Exporter is a Go-based web application that allows users to export AWS GuardDuty findings from multiple US regions into a CSV file. This tool provides a simple web interface for selecting regions and initiating the export process.

## Features
- Web-based interface for easy interaction
- Dynamically fetches and displays available US AWS regions
- Allows selection of multiple regions for export
- Exports GuardDuty findings to a CSV file
- Provides real-time progress updates during the export process

## Prerequisites
- Go 1.16 or later
- AWS account with appropriate permissions to access GuardDuty findings
- AWS CLI configured with valid credentials

## Installation
1. Clone the repository:

git clone https://github.com/yourusername/guardduty-findings-exporter.git

cd guardduty-findings-exporter

2. Install dependencies:

go mod tidy


## Configuration
Ensure your AWS credentials are properly configured. You can do this by setting up the AWS CLI or by setting the appropriate environment variables.

## Usage
1. Start the server:

go run main.go


2. Open a web browser and navigate to `http://localhost:8080`

3. Select the desired US regions from the checkboxes provided

4. Click the "Export Findings" button to start the export process

5. Wait for the export to complete. The application will display the name of the exported CSV file when finished

## File Structure
- `main.go`: The main Go application file
- `index.html`: The HTML template for the web interface

## Contributing
Contributions to improve the GuardDuty Findings Exporter are welcome. Please feel free to submit pull requests or create issues for bugs and feature requests.

## License
[Specify your license here, e.g., MIT, Apache 2.0, etc.]

## Disclaimer
This tool is not officially associated with Amazon Web Services (AWS). Please use it responsibly and in accordance with AWS's terms of service.



