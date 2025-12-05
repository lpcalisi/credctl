### Example: Git Credential Helper

#### Add Provider
````bash
credctl add command gh \
    --command "gh auth token" \
    --login_command "gh auth login --web --clipboard -p https" \
    --run-login \
    --format escaped \
    --template 'protocol=https\nhost=github.com\nusername=x-access-token\npassword={{.raw}}'
````

#### Configure as Git Credential Helper
````bash
git config --global credential.helper "!credctl get gh"
````

#### Ready to Use with Git!
````bash
git clone https://github.com/<owner>/<repo>.git
````