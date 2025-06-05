package main

import (
	"context"
	"os"
	"strings"

	ghrecon "github.com/anotherhadi/gh-recon/gh-recon"
	"github.com/charmbracelet/log"
	"github.com/google/go-github/v72/github"
	flag "github.com/spf13/pflag"
)

func main() {
	var username string
	var token string
	var onlyCommitsLeak bool
	var fromEmail string
	var deep bool
	var silent bool
	var jsonFile string
	var excludeRepos string
	var maxRepoSize int
	var refresh bool
	var showSource bool
	var maxDistance int

	// FLAGS
	flag.StringVarP(&username, "username", "u", "", "GitHub username to analyze")
	flag.StringVarP(&token, "token", "t", "", "GitHub personal access token (e.g. ghp_...)")
	flag.StringVarP(&fromEmail, "email", "e", "", "Search accounts by email address")
	flag.BoolVarP(
		&deep,
		"deep",
		"d",
		false,
		"Enable deep scan (clone repos, regex search, analyse licenses, etc.)",
	)
	flag.IntVar(
		&maxRepoSize,
		"max-size",
		150,
		"Limit the size of repositories to scan (in MB) (only for deep scan)",
	)
	flag.StringVar(
		&excludeRepos,
		"exclude-repo",
		"",
		"Exclude repos from deep scan (comma-separated list, only for deep scan)",
	)
	flag.BoolVarP(
		&refresh,
		"refresh",
		"r",
		false,
		"Refresh the cache (only for deep scan)",
	)
	flag.BoolVarP(
		&showSource,
		"show-source",
		"s",
		false,
		"Show where the information (authors, emails, etc) were found (only for deep scan)",
	)
	flag.IntVarP(
		&maxDistance,
		"max-distance",
		"m",
		20,
		"Maximum Levenshtein distance for matching usernames & emails (only for deep scan)",
	)
	flag.BoolVarP(
		&onlyCommitsLeak,
		"only-commits",
		"c",
		false,
		"Display only commits with author info",
	)
	flag.BoolVarP(&silent, "silent", "S", false, "Suppress all non-essential output")
	flag.StringVarP(&jsonFile, "json", "j", "", "Write results to specified JSON file")

	// FLAGS SETTINGS
	flag.CommandLine.SetNormalizeFunc(wordSepNormalizeFunc)
	flag.CommandLine.SortFlags = false

	flag.Parse()

	// INITIALIZE RECON OBJECT

	r := &ghrecon.Recon{
		Client: github.NewClient(nil),
		Logger: log.NewWithOptions(os.Stderr, log.Options{
			ReportCaller:    false,
			ReportTimestamp: false,
		}),
		Ctx:         context.Background(),
		Silent:      silent,
		JsonFile:    jsonFile,
		MaxRepoSize: maxRepoSize,
		MaxDistance: maxDistance,
		ShowSource:  showSource,
	}

	// CHECK FLAGS

	if username == "" && fromEmail == "" {
		r.Logger.Fatal(
			"Please provide a username with the --username (-u) flag or an email with the --email (-e) flag",
		)
	} else if username != "" {
		username = strings.TrimPrefix(username, "@")
		if err := ghrecon.ParseUsername(username); err != nil {
			r.Logger.Fatal("Invalid username", "err", err)
		}
	}

	if token == "" {
		r.PrintInfo(
			"INFO",
			"It's recommended to set a Github token for better rate limits. You can set it using the --token (-t) flag.",
		)
	} else {
		r.Client = r.Client.WithAuthToken(token)
	}

	// START

	r.Header()

	if fromEmail != "" {
		emailsInfo := r.Email(fromEmail)
		r.WriteJson(
			map[string]any{
				"Authors": emailsInfo,
			},
		)
		return
	}

	if onlyCommitsLeak {
		commitsInfo := r.Commits(username)
		r.WriteJson(
			map[string]any{
				"Authors": commitsInfo,
			},
		)
		return
	}

	userInfo := r.User(username)
	orgsInfo := r.Orgs(username)
	sshKeysInfo := r.SshKeys(username)
	gpgKeysInfo := r.GpgKeys(username)
	sshSigningKeysInfo := r.SshSigningKeys(username)
	socialsInfo := r.Socials(username)
	closeFriendsInfo := r.CloseFriends(username)
	commitsInfo := r.Commits(username)

	results := map[string]any{
		"User":           userInfo,
		"Orgs":           orgsInfo,
		"SSHKeys":        sshKeysInfo,
		"GPGKeys":        gpgKeysInfo,
		"SSHSigningKeys": sshSigningKeysInfo,
		"Socials":        socialsInfo,
		"Commits":        commitsInfo,
		"CloseFriends":   closeFriendsInfo,
	}

	if deep {
		results["Deep"] = r.Deep(username, excludeRepos, refresh)
	}

	r.WriteJson(results)
}
