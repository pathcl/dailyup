# ADR-001: Azure CLI credential over Personal Access Token

## Status

Accepted

## Context

The tool needs to authenticate with the Azure DevOps REST API. Two main options exist:

- **Personal Access Token (PAT)**: A long-lived secret stored in the config file or environment variable. The user must generate it in the ADO UI, set an expiry, and rotate it manually.
- **Azure CLI credential**: Delegates to an existing `az login` session. The Azure SDK's `azidentity.NewAzureCLICredential` exchanges the local session for a short-lived Bearer token automatically.

Most developers who work with Azure already have the Azure CLI installed and stay logged in via `az login`. PATs require an extra manual step and become a source of leaking secrets if the config file is committed or shared.

## Decision

Use `azidentity.NewAzureCLICredential` from `github.com/Azure/azure-sdk-for-go/sdk/azidentity`. On every run the tool requests a fresh token scoped to the Azure DevOps resource ID (`499b84ac-1321-427f-aa17-267ca6975798`).

## Consequences

- No secret to rotate or store. The config file contains only org/project metadata.
- Requires the Azure CLI to be installed and the user to have run `az login` at least once.
- If the CLI session is expired, the error message tells the user to re-run `az login`.
- Token acquisition adds one network round-trip at startup; this is negligible compared to the subsequent API calls.
- PAT support (`NewClientWithToken`) is retained in the client for testing purposes but not exposed via any flag.
