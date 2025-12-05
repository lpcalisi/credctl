### Example: Chainguard APK Registry Authentication

This example shows how to configure `credctl` to authenticate with Chainguard's APK registry (`apk.cgr.dev`) using `chainctl` tokens.

#### Prerequisites

- [chainctl](https://edu.chainguard.dev/chainguard/administration/how-to-install-chainctl/) installed and configured
- SSH access to the remote host (for socket forwarding)

#### Add Provider

```bash
credctl add command chainctl \
    --command "chainctl auth token --audience apk.cgr.dev" \
    --template "export HTTP_AUTH=basic:apk.cgr.dev:user:{{.raw}}"
```

#### Forward Socket to Remote Host

Forward the credctl agent socket via SSH to use credentials on a remote machine:

```bash
ssh -R /tmp/credctl.sock:$HOME/.credctl/agent-readonly.sock <remote-host>
```

> **Note:** Replace `<remote-host>` with your actual SSH target.

#### Ready to Use on Remote Host!

On the remote machine, configure and use the credentials:

```bash
# Point to the forwarded socket
export CREDCTL_SOCK=/tmp/credctl.sock

# Load the authentication credentials
eval $(credctl get chainctl)

# Now you can use apk with Chainguard's registry
apk update
```

Expected output:
```
 [https://apk.cgr.dev/chainguard]
OK: 145906 distinct packages available
```
