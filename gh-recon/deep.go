package ghrecon

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/google/go-github/v72/github"
)

type AuthorOccurrence struct {
	Name    string
	Email   string
	FoundIn []string
}

type EmailOccurrence struct {
	Email   string
	FoundIn []string
}

type DeepResult struct {
	Repositories []Repositorie
	Authors      []AuthorOccurrence
	Emails       []EmailOccurrence
}

type Repositorie struct {
	Repository string
	Owner      string
	Name       string
	Size       int
}

func findEmailsAndOccurrencesInDir(rootPath string) ([]EmailOccurrence, error) {
	emailLocations := make(map[string]map[string]bool)
	emailRegex := regexp.MustCompile(`[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}`)
	normalizedRootPath := filepath.Clean(rootPath)

	err := filepath.WalkDir(rootPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			fmt.Printf("Can't access %s: %v\n", path, err)
			return err
		}
		if !d.IsDir() {
			if strings.Contains(path, ".git/logs/") {
				return nil
			}
			content, err := os.ReadFile(path)
			if err != nil {
				fmt.Printf("Can't read %s: %v\n", path, err)
				return nil
			}

			currentFileEmails := emailRegex.FindAllString(string(content), -1)
			if len(currentFileEmails) > 0 {
				relativePath, errRel := filepath.Rel(normalizedRootPath, path)
				if errRel != nil {
					fmt.Printf("Can't find the relative path %s: %v\n", path, errRel)
					relativePath = path
				}

				for _, email := range currentFileEmails {
					if len(email) > 12 {
						if _, ok := emailLocations[email]; !ok {
							emailLocations[email] = make(map[string]bool)
						}
						emailLocations[email][relativePath] = true
					}
				}
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	var results []EmailOccurrence
	for email, pathSet := range emailLocations {
		var paths []string
		for path := range pathSet {
			paths = append(paths, path)
		}
		results = append(results, EmailOccurrence{Email: email, FoundIn: paths})
	}

	return results, nil
}

func (r Recon) Deep(username, excludeRepos string, refresh bool) (response DeepResult) {
	repositories := []Repositorie{}
	excludeReposList := strings.Split(excludeRepos, ",")
	repos, resp, err := r.Client.Repositories.ListByUser(
		r.Ctx,
		username,
		&github.RepositoryListByUserOptions{
			Type: "all",
		},
	)
	if err != nil {
		r.Logger.Error("Failed to fetch repositories", "err", err)
		return
	}

	r.PrintTitle("📦 Repositories")
	if len(repos) == 0 {
		r.PrintInfo("INFO", "No repositories found")
	} else {
		for _, repo := range repos {
			if slices.Contains(excludeReposList, repo.GetName()) ||
				slices.Contains(excludeReposList, repo.GetOwner().GetLogin()+"/"+repo.GetName()) {
				continue
			}

			maxRepoSize := r.MaxRepoSize * 1024

			if repo.GetSize() > maxRepoSize {
				r.PrintInfo(
					"INFO",
					"Skipping repository "+repo.GetOwner().GetLogin()+"/"+repo.GetName()+" due to size", fmt.Sprintf(
						"%d",
						repo.GetSize()/1024,
					)+"MB > "+fmt.Sprintf(
						"%d",
						maxRepoSize/1024,
					)+"MB",
				)
				continue
			}

			repositories = append(repositories, Repositorie{
				Repository: repo.GetCloneURL(),
				Owner:      repo.GetOwner().GetLogin(),
				Name:       repo.GetName(),
				Size:       repo.GetSize(),
			})
		}
	}
	WaitForRateLimit(resp)

	cmd := exec.Command("git", "--version")
	if err := cmd.Run(); err != nil {
		r.PrintInfo("ERROR", "Git is not installed, please install it to use this feature")
		return
	}

	tmp_folder := "/tmp/ghrecon-" + username

	if folderExists(tmp_folder) {
		if refresh {
			r.PrintInfo("INFO", "Deleting existing folder "+tmp_folder)
			err := os.RemoveAll(tmp_folder)
			if err != nil {
				r.PrintInfo("ERROR", "Failed to delete existing folder "+tmp_folder)
			}
		}
	}

	for _, repo := range repositories {
		destination := tmp_folder + "/" + repo.Owner + "/" + repo.Name
		if folderExists(destination) {
			r.PrintInfo("INFO", "Directory already downloaded, skipping "+repo.Owner+"/"+repo.Name)
			continue
		}

		r.PrintInfo(
			"Downloading",
			repo.Owner+"/"+repo.Name,
			fmt.Sprintf("%d", repo.Size/1024)+"MB",
		)

		cmd := exec.Command(
			"git",
			"clone",
			repo.Repository,
			destination,
		)
		err := cmd.Run()
		if err != nil {
			r.Logger.Error(
				"ERROR",
				"Failed to clone repository",
				"err",
				err,
				"repo",
				repo.Repository,
			)
			continue
		}
	}
	r.PrintInfo("INFO", "Cloned all repositories to "+tmp_folder)

	authorOccurrences := []AuthorOccurrence{}
	mapAuthorToIndex := make(map[string]int)
	for _, repo := range repositories {
		destination := tmp_folder + "/" + repo.Owner + "/" + repo.Name
		if !folderExists(filepath.Join(destination, ".git")) {
			r.Logger.Error(
				"No .git directory found, cannot run git log.",
				"repo",
				repo.Owner+"/"+repo.Name,
				"path",
				destination,
			)
		} else {
			gitLogCmd := exec.Command("git", "log", "--all", "--format=%aN <%aE>")
			gitLogCmd.Dir = destination
			logOutput, logErr := gitLogCmd.Output()

			if logErr != nil {
				if exitErr, ok := logErr.(*exec.ExitError); ok {
					r.Logger.Error("Failed to execute git log (ExitError)", "repo", repo.Owner+"/"+repo.Name, "stderr", string(exitErr.Stderr), "err", logErr)
				} else {
					r.Logger.Error("Failed to execute git log", "repo", repo.Owner+"/"+repo.Name, "err", logErr)
				}
			} else {
				lines := strings.Split(string(logOutput), "\n")
				repoIdentifier := repo.Owner + "/" + repo.Name

				for _, line := range lines {
					trimmedLine := strings.TrimSpace(line)
					if trimmedLine == "" {
						continue
					}

					if index, exists := mapAuthorToIndex[trimmedLine]; exists {
						isRepoListed := false
						for _, foundRepo := range authorOccurrences[index].FoundIn {
							if foundRepo == repoIdentifier {
								isRepoListed = true
								break
							}
						}
						if !isRepoListed {
							authorOccurrences[index].FoundIn = append(authorOccurrences[index].FoundIn, repoIdentifier)
							slices.Sort(authorOccurrences[index].FoundIn)
						}
					} else {
						parts := strings.SplitN(trimmedLine, " <", 2)
						var authorName, authorEmail string
						if len(parts) == 2 {
							authorName = parts[0]
							authorEmail = strings.TrimSuffix(parts[1], ">")
						} else if len(parts) == 1 {
							authorName = "-"
							authorEmail = strings.TrimPrefix(strings.TrimSuffix(parts[0], ">"), "<")
						} else {
							r.Logger.Error("Malformed author line from git log", "line", trimmedLine, "repo", repoIdentifier)
							continue
						}

						authorOccurrences = append(authorOccurrences, AuthorOccurrence{
							Name:    authorName,
							Email:   authorEmail,
							FoundIn: []string{repoIdentifier},
						})
						mapAuthorToIndex[trimmedLine] = len(authorOccurrences) - 1
					}
				}
			}
		}
	}
	for _, author := range authorOccurrences {
		if r.ShowSource {
			r.PrintInfo(
				"Author",
				author.Name+" - "+author.Email,
				"found in:"+strings.Join(author.FoundIn, ", "),
			)
		} else {
			r.PrintInfo(
				"Author",
				author.Name+" - "+author.Email,
			)
		}
	}

	r.PrintInfo("INFO", "Now searching for emails in cloned repositories, this may take a while...")
	emails, err := findEmailsAndOccurrencesInDir(tmp_folder)
	if err != nil {
		r.Logger.Error("Failed to find emails in directory", "err", err)
		return
	}

	if len(emails) == 0 {
		r.PrintInfo("INFO", "No emails found")
	} else {
		r.PrintInfo("INFO", "Found emails:")
		for _, email := range emails {
			if r.ShowSource {
				r.PrintInfo("Email", email.Email, "found in:"+strings.Join(email.FoundIn, ", "))
			} else {
				r.PrintInfo("Email", email.Email)
			}
		}
	}

	response.Repositories = repositories
	response.Authors = authorOccurrences
	response.Emails = emails

	r.PrintNewline()
	return
}
