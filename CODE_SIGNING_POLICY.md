# Code signing policy

The project intends to use the free Open Source code-signing service provided by SignPath.io with a certificate provided by SignPath Foundation, subject to project acceptance.

## Provenance

Official signed releases must be produced from this public repository using the checked-in build workflow. Binary artifacts produced from unpublished source code are not eligible for official signing.

## Roles

- Committers and reviewers: `https://github.com/MohamedTaltlo/CPEFinder` maintainers
- Approvers: `https://github.com/MohamedTaltlo/CPEFinder` owners/maintainers authorized to approve releases

## Release approval

Every release submitted for signing requires manual approval. Changes to source code, build scripts, workflows, signing configuration, or release packaging are reviewed like application code.

## Privacy

See [PRIVACY.md](PRIVACY.md). The program will not transfer information to other networked systems unless specifically requested by the user or operator.

## Security scope

CPEFinder is a device-discovery and inventory utility. It is not intended to identify exploitable vulnerabilities, bypass access controls, crack credentials, or exploit devices.
