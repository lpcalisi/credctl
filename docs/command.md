# Command Provider

The `command` provider executes a local command and returns its output as a credential.

## Usage

```bash
credctl add command <name> --command "<command>"
```

The command:
- Runs on your local machine
- Output (stdout) is returned as the credential
- Result is cached in memory until daemon restart

## Examples

### GitHub CLI
```bash
credctl add command gh --command "gh auth token"
credctl get gh
```

### AWS CLI
```bash
credctl add command aws --command "aws sts get-session-token --output text --query 'Credentials.SessionToken'"
credctl get aws
```

### Custom script
```bash
credctl add command mytoken --command "/path/to/script.sh"
credctl get mytoken
```

## Notes

- Commands execute with your user's environment variables
- Non-zero exit codes return as errors
- Results are cached in memory until daemon restart
