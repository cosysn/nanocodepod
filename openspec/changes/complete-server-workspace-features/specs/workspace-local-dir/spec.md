## ADDED Requirements

### Requirement: Use local directory as workspace source
The server SHALL copy files from a local directory into the workspace when local_dir is provided in the workspace creation request.

#### Scenario: Copy local directory contents
- **WHEN** user creates workspace with local_dir="/home/user/myproject"
- **THEN** server copies all contents from myproject into the workspace directory

#### Scenario: Local directory does not exist
- **WHEN** user provides local_dir that does not exist on server
- **THEN** server returns error indicating directory not found

#### Scenario: Preserve file permissions
- **WHEN** user copies local directory with executable scripts
- **THEN** server preserves executable permissions on copied files
