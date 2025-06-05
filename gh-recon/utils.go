package ghrecon

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/google/go-github/v72/github"
)

var (
	Grey  = lipgloss.Color("#7d7d7d")
	Green = lipgloss.Color("#a6e3a1")
	Red   = lipgloss.Color("#f38ba8")

	GreyStyle  = lipgloss.NewStyle().Foreground(Grey)
	GreenStyle = lipgloss.NewStyle().Foreground(Green)
	RedStyle   = lipgloss.NewStyle().Foreground(Red)
)

func (r Recon) Header() {
	if r.Silent {
		return
	}
	asciiArt := "        __                       \n  ___ _/ /  _______ _______  ___ \n / _ `/ _ \\/ __/ -_) __/ _ \\/ _ \\\n \\_, /_//_/_/  \\__/\\__/\\___/_//_/\n/___/                            "
	fmt.Println(
		GreyStyle.Render(lipgloss.JoinVertical(lipgloss.Right, asciiArt, "@anotherhadi\n")),
	)
}

func ParseUsername(username string) error {
	if username == "" {
		return fmt.Errorf("username is required")
	}
	if strings.Contains(username, " ") {
		return fmt.Errorf("username cannot contain spaces")
	}
	if strings.Contains(username, "@") {
		return fmt.Errorf("username cannot contain @")
	}
	return nil
}

func FetchGitHubAPI(github *github.Client, token, path string) ([]byte, error) {
	url := "https://api.github.com" + path
	userAgent := "GHRecon/1.0"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request for %s: %w", url, err)
	}

	if token != "" {
		req.Header.Set("Authorization", "token "+token)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", userAgent)

	resp, err := github.Client().Do(req)
	if err != nil {
		return nil, fmt.Errorf("error executing request for %s: %w", url, err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf(
			"request for %s failed with status %d: %s",
			url,
			resp.StatusCode,
			string(bodyBytes),
		)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf(
			"error reading response body for %s: %w",
			url,
			err,
		)
	}

	return bodyBytes, nil
}

func (r Recon) PrintNewline() {
	if r.Silent {
		return
	}
	fmt.Println()
}

func (r Recon) PrintTitle(title string) {
	if r.Silent {
		return
	}
	style := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7287fd"))
	fmt.Println(style.Render(title) + "\n")
}

func (r Recon) PrintInfo(key, value string, more ...string) {
	if r.Silent {
		return
	}
	if value == "" || value == "0001-01-01 00:00:00 +0000 UTC" {
		return
	}
	if strings.HasSuffix(key, "n°") {
		fmt.Printf(" %s %s", GreyStyle.Render(key), value)
	} else {
		fmt.Printf(" %s %s", GreyStyle.Render(key+":"), value)
	}
	if len(more) > 0 {
		fmt.Printf(" %s", GreyStyle.Render("("+strings.Join(more, ", ")+")"))
	}
	fmt.Println()
}

func WaitForRateLimit(resp *github.Response) {
	if resp.Rate.Remaining == 0 {
		log.Info(
			"Rate limit reached, waiting for reset... (time:" + resp.Rate.Reset.Time.String() + ")",
		)
		time.Sleep(time.Until(resp.Rate.Reset.Time) + time.Second)
	}
}

func SkipResult(name, email string) bool {
	if name == "github-actions[bot]" || name == "github-actions" {
		return true
	}
	if email == "github-actions[bot]@users.noreply.github.com" ||
		email == "github-actions@github.com" {
		return true
	}
	return false
}

func (r Recon) WriteJson(data any) {
	if r.JsonFile == "" {
		return
	}
	file, err := os.Create(r.JsonFile)
	if err != nil {
		r.Logger.Error("Failed to create JSON file", "err", err)
		return
	}
	defer func() {
		_ = file.Close()
	}()
	as_json, _ := json.MarshalIndent(data, "", "\t")
	_, err = file.Write(as_json)
	if err != nil {
		r.Logger.Error("Failed to write to JSON file", "err", err)
		return
	}
	r.PrintInfo("INFO", "JSON file created successfully", "file", r.JsonFile)
}

func folderExists(path string) bool {
	if stat, err := os.Stat(path); err == nil && stat.IsDir() {
		return true
	}
	return false
}

func levenshteinDistance(s1, s2 string) int {
	len1 := len(s1)
	len2 := len(s2)

	dp := make([][]int, len1+1)
	for i := range dp {
		dp[i] = make([]int, len2+1)
	}

	for i := 0; i <= len1; i++ {
		dp[i][0] = i
	}

	for j := 0; j <= len2; j++ {
		dp[0][j] = j
	}

	for i := 1; i <= len1; i++ {
		for j := 1; j <= len2; j++ {
			cost := 0
			if s1[i-1] != s2[j-1] {
				cost = 1
			}

			dp[i][j] = int(
				math.Min(
					float64(dp[i-1][j]+1),
					math.Min(float64(dp[i][j-1]+1), float64(dp[i-1][j-1]+cost)),
				),
			)
		}
	}

	return dp[len1][len2]
}
