package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var (
	forceMonorepo bool
	intervalMins  string
)

func init() {
	flag.BoolVar(&forceMonorepo, "mr", false, "Force monorepo mode (auto-detects if not set)")
	flag.BoolVar(&forceMonorepo, "monorepo", false, "Force monorepo mode (auto-detects if not set)")
	flag.StringVar(&intervalMins, "i", "0.5", "Check interval in minutes (0.5-30)")
	flag.StringVar(&intervalMins, "interval", "0.5", "Check interval in minutes (0.5-30)")

	flag.Usage = showHelp
}

func showHelp() {
	fmt.Println("🚀 Git Air - Automatic Git synchronization service")
	fmt.Println("\nUSAGE:")
	fmt.Println("  git-air [options]")
	fmt.Println("\nOPTIONS:")
	fmt.Println("  -h, --help              Show this help screen")
	fmt.Println("  -i, --interval <mins>   Check interval in minutes (0.5-30)")
	fmt.Println("                          Examples: 0.5, 1, 2, 5, 10, 30")
	fmt.Println("                          Default: 0.5 (30 seconds)")
	fmt.Println("  -mr, --monorepo         Force monorepo mode")
	fmt.Println("                          (auto-detects if not set)")
	fmt.Println("\nEXAMPLES:")
	fmt.Println("  git-air                 # Run with default 30 second interval")
	fmt.Println("  git-air -i 1            # Check every 1 minute")
	fmt.Println("  git-air -i 5 -mr        # Check every 5 minutes, force monorepo")
	fmt.Println("  git-air --interval 10   # Check every 10 minutes")
	fmt.Println("\nDESCRIPTION:")
	fmt.Println("  Automatically discovers and synchronizes all Git repositories")
	fmt.Println("  in the current directory and subdirectories.")
	fmt.Println("\n  Features:")
	fmt.Println("  • Auto-commits changes with timestamp")
	fmt.Println("  • Pushes to ALL configured remotes")
	fmt.Println("  • Pulls updates for inter-project communication")
	fmt.Println("  • Handles monorepos with submodules")
	fmt.Println()
}

func parseInterval(intervalStr string) (time.Duration, error) {
	mins, err := strconv.ParseFloat(intervalStr, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid interval format: %v", err)
	}

	if mins < 0.5 || mins > 30 {
		return 0, fmt.Errorf("interval must be between 0.5 and 30 minutes, got: %.1f", mins)
	}

	return time.Duration(mins * float64(time.Minute)), nil
}

func main() {
	flag.Parse()

	// Parse and validate interval
	checkInterval, err := parseInterval(intervalMins)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error: %v\n\n", err)
		showHelp()
		os.Exit(1)
	}

	fmt.Println("🚀 Git Air - Auto sync all Git repos")
	fmt.Println("📡 Inter-project communication via Git synchronization")
	fmt.Println("📚 Supports monorepos and multi-repos")
	fmt.Printf("⏱️  Check interval: %.1f minutes\n", checkInterval.Minutes())
	if forceMonorepo {
		fmt.Println("🔧 Monorepo mode: FORCED")
	} else {
		fmt.Println("🔧 Monorepo mode: AUTO-DETECT")
	}
	fmt.Println()

	// Find all git repos in current directory and subdirs
	repos, err := findGitRepos(".")
	if err != nil {
		log.Fatalf("❌ Error finding repositories: %v\n", err)
	}

	if len(repos) == 0 {
		fmt.Println("⚠️  No Git repositories found in current directory")
		fmt.Println("💡 Make sure you're in a directory containing Git repositories")
		os.Exit(0)
	}

	fmt.Printf("Found %d Git repositories\n", len(repos))
	for _, repo := range repos {
		repoType := "repo"
		if forceMonorepo || isMonorepo(repo) {
			repoType = "MONOREPO"
		}
		fmt.Printf("  📁 %s [%s]\n", repo, repoType)
	}
	fmt.Println()

	// Calculate pull interval (every minute or every checkInterval, whichever is longer)
	pullInterval := time.Minute
	if checkInterval > pullInterval {
		pullInterval = checkInterval
	}

	// Main loop
	lastPull := time.Now()
	iteration := 0

	for {
		iteration++
		fmt.Printf("🔄 Check cycle #%d\n", iteration)

		// Auto commit and push changes
		changesFound := false
		for _, repo := range repos {
			if processRepo(repo, forceMonorepo) {
				changesFound = true
			}
		}

		if !changesFound {
			fmt.Println("  ✓ No changes detected")
		}

		// Pull from all repos at pull interval
		if time.Since(lastPull) >= pullInterval {
			fmt.Println("\n📡 Checking for inter-project updates...")
			for _, repo := range repos {
				pullUpdates(repo)
			}
			lastPull = time.Now()
		}

		fmt.Printf("\n💤 Sleeping for %.1f minutes...\n\n", checkInterval.Minutes())
		time.Sleep(checkInterval)
	}
}

// findGitRepos finds all .git directories
func findGitRepos(root string) ([]string, error) {
	var repos []string
	
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}
		
		// Skip some common dirs
		if info.IsDir() && (info.Name() == "node_modules" || info.Name() == "vendor") {
			return filepath.SkipDir
		}
		
		// Found a .git directory
		if info.IsDir() && info.Name() == ".git" {
			repoPath := filepath.Dir(path)
			repos = append(repos, repoPath)
			return filepath.SkipDir // Don't go into .git
		}
		
		return nil
	})
	
	return repos, err
}

// processRepo handles one git repository, returns true if changes were committed
func processRepo(repoPath string, forceMonorepo bool) bool {
	// Change to repo directory
	oldDir, err := os.Getwd()
	if err != nil {
		fmt.Printf("  ❌ Error getting working directory: %v\n", err)
		return false
	}

	if err := os.Chdir(repoPath); err != nil {
		fmt.Printf("  ❌ Error changing to %s: %v\n", repoPath, err)
		return false
	}
	defer os.Chdir(oldDir)

	// Determine if this is a monorepo
	isMonorepoMode := forceMonorepo || isMonorepo(repoPath)

	// For monorepos: sync submodules FIRST
	if isMonorepoMode {
		if !syncSubmodules(repoPath) {
			fmt.Printf("  ❌ Skipping %s - submodule sync failed\n", filepath.Base(repoPath))
			return false
		}
	}

	// Check if there are changes AFTER submodule sync
	if !hasChanges() {
		return false // No changes to commit
	}

	repoName := filepath.Base(repoPath)
	repoType := ""
	if isMonorepoMode {
		repoType = " [MONOREPO]"
	}
	fmt.Printf("📝 %s%s: Auto committing changes...\n", repoName, repoType)

	// Auto commit with monorepo-aware message
	if !runGit("add", ".") {
		fmt.Printf("  ❌ Error staging changes in %s\n", repoName)
		return false
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	commitMsg := "auto commit - " + timestamp
	if isMonorepoMode {
		commitMsg = "auto commit (monorepo) - " + timestamp
	}

	if !runGit("commit", "-m", commitMsg) {
		fmt.Printf("  ⚠️  Commit failed in %s (may be empty or have errors)\n", repoName)
		return false
	}

	fmt.Printf("  ✓ Committed changes in %s\n", repoName)

	// Push to all remotes immediately
	pushToAllRemotes()

	return true
}

// pullUpdates pulls from remotes for inter-project communication
func pullUpdates(repoPath string) {
	// Change to repo directory
	oldDir, err := os.Getwd()
	if err != nil {
		fmt.Printf("  ❌ Error getting working directory: %v\n", err)
		return
	}

	if err := os.Chdir(repoPath); err != nil {
		fmt.Printf("  ❌ Error changing to %s: %v\n", repoPath, err)
		return
	}
	defer os.Chdir(oldDir)

	pullFromRemotes()
}

// hasChanges checks if repo has uncommitted changes
func hasChanges() bool {
	cmd := exec.Command("git", "status", "--porcelain")
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	return len(strings.TrimSpace(string(output))) > 0
}

// pushToAllRemotes pushes to all configured remotes
func pushToAllRemotes() {
	remotes := getRemotes()
	if len(remotes) == 0 {
		fmt.Println("  ⚠️  No remotes configured, skipping push")
		return
	}

	branch := getCurrentBranch()
	successCount := 0
	for _, remote := range remotes {
		fmt.Printf("  🚀 Pushing to %s...", remote)
		if runGit("push", remote, branch) {
			fmt.Printf(" ✓\n")
			successCount++
		} else {
			fmt.Printf(" ❌ failed\n")
		}
	}

	if successCount > 0 {
		fmt.Printf("  ✓ Successfully pushed to %d/%d remotes\n", successCount, len(remotes))
	}
}

// pullFromRemotes pulls from remotes for inter-project communication
func pullFromRemotes() {
	remotes := getRemotes()
	if len(remotes) == 0 {
		return
	}

	branch := getCurrentBranch()
	repoName := filepath.Base(getCurrentDir())

	// Try to pull from each remote
	for _, remote := range remotes {
		fmt.Printf("  📥 %s: Checking %s for updates...", repoName, remote)
		if !runGit("fetch", remote) {
			fmt.Printf(" ❌ fetch failed\n")
			continue
		}

		// Check if there are remote changes
		if hasRemoteChanges(remote, branch) {
			fmt.Printf("\n  📡 %s: Pulling updates from %s...", repoName, remote)
			if runGit("pull", remote, branch) {
				fmt.Printf(" ✓\n")
			} else {
				fmt.Printf(" ❌ pull failed\n")
			}
		} else {
			fmt.Printf(" ✓ up to date\n")
		}
	}
}

// getRemotes returns list of remote names
func getRemotes() []string {
	cmd := exec.Command("git", "remote")
	output, err := cmd.Output()
	if err != nil {
		return []string{}
	}
	
	remotes := strings.Fields(string(output))
	return remotes
}

// getCurrentBranch returns current branch name
func getCurrentBranch() string {
	cmd := exec.Command("git", "branch", "--show-current")
	output, err := cmd.Output()
	if err != nil {
		return "main" // fallback
	}
	return strings.TrimSpace(string(output))
}

// runGit runs a git command and returns success
func runGit(args ...string) bool {
	cmd := exec.Command("git", args...)
	err := cmd.Run()
	if err != nil {
		return false
	}
	return true
}

// hasRemoteChanges checks if remote has changes
func hasRemoteChanges(remote, branch string) bool {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	localOut, err := cmd.Output()
	if err != nil {
		return false
	}
	
	cmd = exec.Command("git", "rev-parse", remote+"/"+branch)
	remoteOut, err := cmd.Output()
	if err != nil {
		return false
	}
	
	return string(localOut) != string(remoteOut)
}

// getCurrentDir returns current directory
func getCurrentDir() string {
	dir, _ := os.Getwd()
	return dir
}

// isMonorepo checks if a repository contains submodules or nested repos
func isMonorepo(repoPath string) bool {
	// Check for .gitmodules file (Git submodules)
	gitmodules := filepath.Join(repoPath, ".gitmodules")
	if _, err := os.Stat(gitmodules); err == nil {
		return true
	}
	
	// Check for nested .git directories (indicates multiple projects)
	nestedRepos := 0
	filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() && info.Name() == ".git" && path != filepath.Join(repoPath, ".git") {
			nestedRepos++
			if nestedRepos > 0 {
				return filepath.SkipDir // Found nested repos, it's a monorepo
			}
		}
		return nil
	})
	
	return nestedRepos > 0
}

// syncSubmodules ensures all submodules are updated before main repo commit
func syncSubmodules(repoPath string) bool {
	// Change to repo directory
	oldDir, err := os.Getwd()
	if err != nil {
		fmt.Printf("  ❌ Error getting working directory: %v\n", err)
		return false
	}

	if err := os.Chdir(repoPath); err != nil {
		fmt.Printf("  ❌ Error changing to %s: %v\n", repoPath, err)
		return false
	}
	defer os.Chdir(oldDir)

	// Check if there are submodules
	gitmodules := filepath.Join(repoPath, ".gitmodules")
	if _, err := os.Stat(gitmodules); err != nil {
		return true // No submodules, all good
	}

	fmt.Printf("  📦 Syncing submodules...")

	// Update all submodules
	if !runGit("submodule", "update", "--remote", "--merge") {
		fmt.Printf(" ❌ failed\n")
		return false
	}

	// Add any submodule changes
	if !runGit("add", ".") {
		fmt.Printf(" ⚠️  failed to stage submodule changes\n")
		return false
	}

	fmt.Printf(" ✓\n")
	return true
}