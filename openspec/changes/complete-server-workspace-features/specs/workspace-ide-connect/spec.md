## ADDED Requirements

### Requirement: Launch IDE for workspace
The CLI SHALL launch an IDE to connect to the workspace when requested.

#### Scenario: Launch VS Code
- **WHEN** user specifies --ide=vscode after workspace is created
- **THEN** CLI opens vscode://localhost:... URL to connect to the workspace

#### Scenario: Launch JetBrains IDE
- **WHEN** user specifies --ide=jetbrains after workspace is created
- **THEN** CLI opens jetbrains://localhost:... URL to connect to the workspace

#### Scenario: IDE not installed
- **WHEN** user requests IDE launch but IDE is not installed
- **THEN** system shows error suggesting IDE installation

#### Scenario: Connect flag on workspace create
- **WHEN** user creates workspace with --connect flag
- **THEN** IDE launches automatically after workspace is ready
