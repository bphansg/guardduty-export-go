import boto3
import csv
from datetime import datetime

def get_us_regions():
    ec2 = boto3.client('ec2', region_name='us-east-1')
    regions = [region['RegionName'] for region in ec2.describe_regions()['Regions'] if region['RegionName'].startswith('us-')]
    return regions

def get_guardduty_findings(region):
    guardduty = boto3.client('guardduty', region_name=region)
    detector_ids = guardduty.list_detectors()['DetectorIds']
    
    findings = []
    for detector_id in detector_ids:
        paginator = guardduty.get_paginator('list_findings')
        for page in paginator.paginate(DetectorId=detector_id):
            finding_ids = page['FindingIds']
            if finding_ids:
                result = guardduty.get_findings(DetectorId=detector_id, FindingIds=finding_ids)
                findings.extend(result['Findings'])
    
    return findings

def export_to_csv(findings, filename):
    with open(filename, 'w', newline='') as csvfile:
        fieldnames = ['Region', 'FindingId', 'Title', 'Description', 'Severity', 'CreatedAt', 'UpdatedAt']
        writer = csv.DictWriter(csvfile, fieldnames=fieldnames)
        
        writer.writeheader()
        for finding in findings:
            writer.writerow({
                'Region': finding['Region'],
                'FindingId': finding['Id'],
                'Title': finding['Title'],
                'Description': finding['Description'],
                'Severity': finding['Severity'],
                'CreatedAt': finding['CreatedAt'],
                'UpdatedAt': finding['UpdatedAt']
            })

def main():
    regions = get_us_regions()
    all_findings = []

    for region in regions:
        print(f"Fetching GuardDuty findings for region: {region}")
        findings = get_guardduty_findings(region)
        for finding in findings:
            finding['Region'] = region
        all_findings.extend(findings)

    timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
    filename = f"guardduty_findings_{timestamp}.csv"
    export_to_csv(all_findings, filename)
    print(f"Exported {len(all_findings)} findings to {filename}")

if __name__ == "__main__":
    main()
