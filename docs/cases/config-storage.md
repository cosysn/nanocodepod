# Config Storage Test Cases

## TC001: Config Directory Creation

**Preconditions**: None

**Test Steps**:
1. Run `codepod init`

**Expected Result**:
- `~/.codepod/` directory created
- `~/.codepod/workspaces/` created
- `~/.codepod/keys/` created
- `~/.codepod/tools/` created
- `~/.codepod/config.yaml` created with defaults

---

## TC002: Default Config Values

**Preconditions**: None

**Test Steps**:
1. Run `codepod init --force`
2. Run `codepod config list`

**Expected Result**:
- WSL Distribution: Ubuntu-22.04
- Docker Host: tcp://localhost:2375
- Default IDE: vscode
- SSH Port: 2222
- Port Pool: 22000-22999

---

## TC003: Config Set

**Preconditions**: Config initialized

**Test Steps**:
1. Run `codepod config set wsl.distribution Ubuntu-20.04`
2. Run `codepod config get wsl.distribution`

**Expected Result**:
- Output: `Ubuntu-20.04`

---

## TC004: Config Persistence

**Preconditions**: Config modified

**Test Steps**:
1. Run `codepod config set wsl.distribution Ubuntu-22.04`
2. Kill and restart codepod
3. Run `codepod config get wsl.distribution`

**Expected Result**:
- Value persists after restart
- Output: `Ubuntu-22.04`

---

## TC005: Config Reset

**Preconditions**: Config modified

**Test Steps**:
1. Run `codepod config set wsl.distribution Custom`
2. Run `codepod config reset`
3. Run `codepod config get wsl.distribution`

**Expected Result**:
- Returns to default value: `Ubuntu-22.04`

---

## TC006: Invalid Config Key

**Preconditions**: Config initialized

**Test Steps**:
1. Run `codepod config set invalid.key value`

**Expected Result**:
- Error: `unknown config key: invalid.key`
- Exit code: 1
