<div align="center">
    <img src="https://raw.githubusercontent.com/anotherhadi/gh-recon/main/.github/assets/logo.png" width="120px" />
</div>

<br>

# GH-Recon 🔍

<p>
    <a href="https://github.com/anotherhadi/gh-recon/releases"><img src="https://img.shields.io/github/release/anotherhadi/gh-recon.svg" alt="Latest Release"></a>
    <a href="https://pkg.go.dev/github.com/anotherhadi/gh-recon?tab=doc"><img src="https://godoc.org/github.com/anotherhadi/gh-recon?status.svg" alt="GoDoc"></a>
    <a href="https://goreportcard.com/report/github.com/anotherhadi/gh-recon"><img src="https://goreportcard.com/badge/github.com/anotherhadi/gh-recon" alt="GoReportCard"></a>
</p>

## 🧾 Project Overview

Retrieves and aggregates public OSINT data about a GitHub user using Go and the GitHub API.
Finds hidden emails in commit history, previous usernames, friends, other GitHub accounts, and more.

## 🚀 Features

- Retrieve basic user profile information (username, ID, avatar, bio, creation dates)
- List organizations and roles
- Fetch SSH and GPG keys
- Enumerate social accounts
- Extract unique commit authors (name + email)
- Find close friends
- Find Github accounts using an email address
- Export results to JSON
- Deep scan option (clone repositories, regex search, analyze licenses, etc.)

## ⚠️ Disclaimer

This tool is intended for educational purposes only. Use responsibly and ensure you have permission to access the data you are querying.

## 📋 Prerequisites

- Go 1.18+
- GitHub Personal Access Token (recommended for higher rate limits): Create a GitHub API token with no permissions/no scope. This will be equivalent to public GitHub access, but it will allow access to use the GitHub Search API.

## 📦 Installation

### With Go

```bash
go install github.com/anotherhadi/gh-recon@latest
```

### With Nix/NixOS

<details>
<summary>Click to expand</summary>

**From anywhere (using the repo URL):**

```bash
nix run github:anotherhadi/gh-recon -- --username TARGET_USER [--token YOUR_TOKEN]
```

**Permanent Installation:**

```bash
# add the flake to your flake.nix
{
  inputs = {
    gh-recon.url = "github:anotherhadi/gh-recon";
  };
}

# then add it to your packages
environment.systemPackages = with pkgs; [ # or home.packages
  gh-recon
];
```

</details>

## 🧪 Usage

```bash
gh-recon --username TARGET_USER [--token YOUR_TOKEN]
```

### Flags

```txt
  -u, --username string       GitHub username to analyze
  -t, --token string          GitHub personal access token (e.g. ghp_...)
  -e, --email string          Search accounts by email address
  -d, --deep                  Enable deep scan (clone repos, regex search, analyse licenses, etc.)
      --max-size int          Limit the size of repositories to scan (in MB) (only for deep scan) (default 150)
      --exclude-repo string   Exclude repos from deep scan (comma-separated list, only for deep scan)
  -r, --refresh               Refresh the cache (only for deep scan)
  -s, --show-source           Show where the information (authors, emails, etc) were found (only for deep scan)
  -c, --only-commits          Display only commits with author info
  -S, --silent                Suppress all non-essential output
  -j, --json string           Write results to specified JSON file
```

## 💡 Examples

```bash
gh-recon --username anotherhadi --token ghp_ABC123...
gh-recon --email myemail@gmail.com
gh-recon --username anotherhadi --json output.json --deep
```

## 🕵️‍♂️ Cover your tracks

Understanding what information about you is publicly visible is the first step to managing your online presence. gh-recon can help you identify your own publicly available data on GitHub. Here’s how you can take steps to protect your privacy and security:

- **Review your public profile**: Regularly check your GitHub profile and repositories to ensure that you are not unintentionally exposing sensitive information.
- **Manage email exposure**: Use GitHub's settings to control which email addresses are visible on your profile and in commit history. You can also use a no-reply email address for commits. Delete/modify any sensitive information in your commit history.
- **Be Mindful of Repository Content**: Avoid including sensitive information in your repositories, such as API keys, passwords, emails or personal data. Use `.gitignore` to exclude files that contain sensitive information.

You can also use a tool like [TruffleHog](github.com/trufflesecurity/trufflehog) to scan your repositories specifically for exposed secrets and tokens.

**Useful links:**

- [Blocking command line pushes that expose your personal email address](https://docs.github.com/en/account-and-profile/setting-up-and-managing-your-personal-account-on-github/managing-email-preferences/blocking-command-line-pushes-that-expose-your-personal-email-address)
- [No-reply email address](https://docs.github.com/en/account-and-profile/setting-up-and-managing-your-personal-account-on-github/managing-email-preferences/setting-your-commit-email-address)

## 🤝 Contributing

Feel free to contribute! See [CONTRIBUTING.md](CONTRIBUTING.md) for details.
