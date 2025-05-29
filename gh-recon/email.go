package ghrecon

import (
	"fmt"

	"github.com/google/go-github/v72/github"
)

type EmailResult struct {
	Name         string
	Email        string
	Username     string
	Occurrences   int
	FirstFoundIn string
}

func (r Recon) Email(email string) (response []EmailResult) {
	r.PrintTitle("✉️ Email")

	results := make(map[string]EmailResult)

	collect := func(date string) error {
		for page := 1; page <= 10; page++ {
			result, resp, err := r.Client.Search.Commits(
				r.Ctx,
				fmt.Sprintf("author-email:%s author-date:%s", email, date),
				&github.SearchOptions{
					Sort:        "author-date",
					Order:       "desc",
					ListOptions: github.ListOptions{PerPage: 100, Page: page},
				},
			)
			if err != nil {
				return fmt.Errorf("fetch page %d (%s): %w", page, date, err)
			}
			WaitForRateLimit(resp)
			if len(result.Commits) == 0 {
				break
			}
			for _, item := range result.Commits {
				name := item.Commit.GetAuthor().GetName()
				email := item.Commit.GetAuthor().GetEmail()
				login := item.GetAuthor().GetLogin()
				if login == "" {
					login = "Unknown"
				}
				if SkipResult(name, email) {
					continue
				}
				if _, seen := results[name+" - "+email+" - "+login]; !seen {
					author := EmailResult{
						Name:       name,
						Email:      email,
						Username:   login,
						Occurrences: 1,
						FirstFoundIn: item.GetRepository().Owner.GetLogin() + "/" + item.GetRepository().
							GetName(),
					}
					results[name+" - "+email+" - "+login] = author
				} else {
					result := results[name+" - "+email+" - "+login]
					result.Occurrences++
					results[name+" - "+email+" - "+login] = result
				}
			}
		}
		return nil
	}

	// Range of dates to bypass the limit of 1000 results
	for _, date := range []string{
		"<2023-01-01", "2023-01-01..2023-12-31",
		"2024-01-01..2024-05-31",
		"2024-06-01..2024-12-31",
		"2025-01-01..2025-05-31",
		"2025-06-01..2025-12-31",
		">2026-01-01",
	} {
		if err := collect(date); err != nil {
			r.Logger.Error("Failed to fetch commits", "err", err, "date", date)
		}
	}

	for _, result := range results {
		r.PrintInfo(
			"Author",
			result.Name+" - "+result.Email+" - @"+result.Username,
			"first from "+result.FirstFoundIn+" (x"+fmt.Sprint(result.Occurrences)+")",
		)
		response = append(response, result)
	}
	if len(results) == 0 {
		r.PrintInfo("INFO", "No commits found")
	}

	r.PrintNewline()
	return
}
