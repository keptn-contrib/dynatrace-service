# Release Notes 0.9.0

## New Features

- Adding the Problem URL of dynatrace as a label and sending informational and configuration change events for auto-remediation actions #177
- Run distributor as sidecar #175
- Don't exit if API connection check goes wrong #171
- Allow to install dynatrace-sli-service in any namespace #172

## Support production readiness

*Removed two features since they cannot be performed in a production-like environment.*

- Removed check for OneAgent #175
- Removed calculated metric creation #168
