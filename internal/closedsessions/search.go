package closedsessions

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/olekukonko/tablewriter"
)

const sessionPrefix = "restore-session-"

var restoreSessionPattern = regexp.MustCompile(`^restore-session-(\d{8})(.*)$`)

// Result is a relative path split into directory and filename components.
type Result struct {
	Directory string
	Filename  string
}

// Search enumerates files under basePath matching session filenames for the requested date range.
func Search(basePath, dateArg string) ([]Result, error) {
	basePath = strings.TrimSpace(basePath)
	if basePath == "" {
		return nil, fmt.Errorf("base search path is required")
	}

	allowedDates, err := ParseDateRequest(dateArg, time.Now())
	if err != nil {
		return nil, err
	}

	matches := make(map[string]Result)
	allowedSet := make(map[string]struct{}, len(allowedDates))
	for _, d := range allowedDates {
		allowedSet[d.Format("20060102")] = struct{}{}
	}

	walkErr := filepath.WalkDir(basePath, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}

		name := entry.Name()
		if !strings.HasPrefix(name, sessionPrefix) {
			return nil
		}

		match := restoreSessionPattern.FindStringSubmatch(name)
		if len(match) < 2 {
			return nil
		}

		dateStamp := match[1]
		if _, ok := allowedSet[dateStamp]; !ok {
			return nil
		}

		rel, err := filepath.Rel(basePath, path)
		if err != nil {
			return err
		}
		dir := filepath.Dir(rel)
		if dir == "." || dir == "" {
			dir = "."
		}

		result := Result{Directory: dir, Filename: filepath.Base(rel)}
		matches[path] = result
		return nil
	})
	if walkErr != nil {
		return nil, walkErr
	}

	results := make([]Result, 0, len(matches))
	for _, result := range matches {
		results = append(results, result)
	}
	sort.Slice(results, func(i, j int) bool {
		if results[i].Directory == results[j].Directory {
			return results[i].Filename < results[j].Filename
		}
		return results[i].Directory < results[j].Directory
	})

	return results, nil
}

// ParseDateRequest receives either an explicit YYYYMMDD date or a relative duration like 1d, 4d, 1w, 1m.
var relativeDatePattern = regexp.MustCompile(`^(\d+)([dwm])$`)

func ParseDateRequest(dateArg string, now time.Time) ([]time.Time, error) {
	dateArg = strings.TrimSpace(dateArg)
	if dateArg == "" {
		return []time.Time{normalizeDay(now)}, nil
	}

	if len(dateArg) == 8 {
		parsed, err := time.Parse("20060102", dateArg)
		if err != nil {
			return nil, fmt.Errorf("invalid date %q: must be YYYYMMDD", dateArg)
		}
		return []time.Time{normalizeDay(parsed)}, nil
	}

	match := relativeDatePattern.FindStringSubmatch(dateArg)
	if len(match) != 3 {
		return nil, fmt.Errorf("invalid date argument %q: supported values are YYYYMMDD, 1d, 4d, 1w, 1m", dateArg)
	}

	amount, err := strconv.Atoi(match[1])
	if err != nil || amount < 1 {
		return nil, fmt.Errorf("invalid duration %q", dateArg)
	}

	unit := match[2]
	var dates []time.Time
	days := amount
	if unit == "w" {
		days = amount * 7
	}
	if unit == "m" {
		days = amount * 30
	}

	for i := 0; i < days; i++ {
		dates = append(dates, normalizeDay(now.AddDate(0, 0, -i)))
	}

	return removeDuplicateDates(dates), nil
}

func normalizeDay(in time.Time) time.Time {
	return time.Date(in.Year(), in.Month(), in.Day(), 0, 0, 0, 0, time.UTC)
}

func removeDuplicateDates(values []time.Time) []time.Time {
	seen := make(map[string]struct{}, len(values))
	unique := make([]time.Time, 0, len(values))
	for _, value := range values {
		key := value.Format("20060102")
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		unique = append(unique, value)
	}
	return unique
}

// RenderTable formats the results as a 2-column table with the base path stripped.
func RenderTable(results []Result) string {
	if len(results) == 0 {
		return ""
	}

	builder := &strings.Builder{}
	table := tablewriter.NewWriter(builder)
	table.Header("DIRECTORY", "FILENAME")

	for _, result := range results {
		table.Append(result.Directory, result.Filename)
	}

	_ = table.Render()
	return builder.String()
}

// ConfigDir returns the directory where the config file is stored.
func ConfigDir() string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return filepath.Join(".", ".config", "closed-sessions")
	}
	return filepath.Join(home, ".config", "closed-sessions")
}

// ConfigFile returns the YAML config file path for closed-sessions.
func ConfigFile() string {
	return filepath.Join(ConfigDir(), "config.yaml")
}

// DefaultBaseSearchPath returns the default search root used if no config value is set.
func DefaultBaseSearchPath() string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return filepath.Join(".", "Work")
	}
	return filepath.Join(home, "Work")
}
