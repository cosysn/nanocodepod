## ADDED Requirements

### Requirement: Build custom Docker image from devcontainer.json
The server SHALL parse .devcontainer.json and build a custom Docker image when devcontainer_path is provided.

#### Scenario: Parse valid devcontainer.json
- **WHEN** user provides devcontainer_path to a valid .devcontainer.json
- **THEN** server reads the configuration and extracts image or dockerfile details

#### Scenario: Build image from Dockerfile
- **WHEN** devcontainer.json specifies a Dockerfile
- **THEN** server builds the Docker image using the specified Dockerfile

#### Scenario: Use prebuilt image from config
- **WHEN** devcontainer.json specifies image directly
- **THEN** server pulls or uses the specified image directly

#### Scenario: Build fails
- **WHEN** Docker build fails due to errors in Dockerfile
- **THEN** server returns error with build output for debugging

#### Scenario: No devcontainer.json at path
- **WHEN** user provides devcontainer_path that does not exist
- **THEN** server returns error indicating file not found
