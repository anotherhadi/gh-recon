package ghrecon

import (
	"fmt"

	"github.com/google/go-github/v72/github"
)

type CommitsResult struct {
	Name         string
	Email        string
	Occurrences   int
	FirstFoundIn string
}

func (r Recon) Commits(username string) (response []CommitsResult) {
	r.PrintTitle("🐙 Commits")

	results := make(map[string]CommitsResult)

	collect := func(date string) error {
		for page := 1; page <= 10; page++ {
			result, resp, err := r.Client.Search.Commits(
				r.Ctx,
				fmt.Sprintf("author:%s author-date:%s", username, date),
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
				if SkipResult(name, email) {
					continue
				}
				if _, seen := results[name+" - "+email]; !seen {
					author := CommitsResult{
						Name:       name,
						Email:      email,
						Occurrences: 1,
						FirstFoundIn: item.GetRepository().Owner.GetLogin() + "/" + item.GetRepository().
							GetName(),
					}
					results[name+" - "+email] = author
				} else {
					result := results[name+" - "+email]
					result.Occurrences++
					results[name+" - "+email] = result
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
		if r.ShowSource {
			r.PrintInfo(
				"Author",
				result.Name+" - "+result.Email,
				"first from "+result.FirstFoundIn+" (x"+fmt.Sprint(result.Occurences)+")",
			)
		} else {
			r.PrintInfo(
				"Author",
				result.Name+" - "+result.Email,
			)
		}
		response = append(response, result)
	}
	if len(results) == 0 {
		r.PrintInfo("INFO", "No commits found")
	}
	r.PrintNewline()

	return
}
