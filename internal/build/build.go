package build

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/alwaysgolang/hippo-cli/internal/ui"
	"github.com/alwaysgolang/hippo-cli/templates"
	"github.com/fatih/color"
	"github.com/schollz/progressbar/v3"
)

const templateModuleName = "gotemplate"

type Options struct {
	Verbose   bool
	Cinematic bool
	Delay     time.Duration
}

func Run(opts Options) error {
	ui.Banner()

	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	serviceName := filepath.Base(wd)

	color.Cyan("🚀 Building service: %s\n", serviceName)

	needGoInit := !hasGoMod(wd)
	needGitInit := !hasGitRepo(wd)

	// базовые шаги: copy + env + tidy
	steps := 3
	if needGoInit {
		steps++
	}
	if needGitInit {
		steps++
	}

	// cinematic delay
	delay := time.Duration(0)
	if opts.Cinematic {
		delay = 550 * time.Millisecond
		if opts.Delay > 0 {
			delay = opts.Delay
		}
	}

	bar := progressbar.NewOptions(steps,
		progressbar.OptionSetWriter(os.Stderr), // важно: бар в stderr
		progressbar.OptionSetWidth(40),
		progressbar.OptionSetDescription("Progress"),
		progressbar.OptionShowCount(),
	)

	// 1) copy
	if err := runStep("Copying template...", func() error {
		return copyFromEmbed("rest", wd, serviceName)
	}, delay); err != nil {
		return err
	}
	_ = bar.Add(1)

	// 2) env
	if err := runStep("Creating environment...", func() error {
		return createEnvIfNotExists(wd)
	}, delay); err != nil {
		return err
	}
	_ = bar.Add(1)

	// 3) go mod init (only if needed)
	if needGoInit {
		if err := runStep("Initializing go module...", func() error {
			_, err := runCmd(wd, opts.Verbose, "go", "mod", "init", serviceName)
			return err
		}, delay); err != nil {
			return err
		}
		_ = bar.Add(1)
	} else {
		color.Yellow("✔ go.mod already exists")
	}

	// 4) tidy (always)
	if err := runStep("Running go mod tidy...", func() error {
		out, err := runCmd(wd, opts.Verbose, "go", "mod", "tidy")
		if err != nil && !opts.Verbose && len(out) > 0 {
			fmt.Println(string(out))
		}
		return err
	}, delay); err != nil {
		return err
	}
	_ = bar.Add(1)

	// 5) git init + commit (only if needed)
	if needGitInit {
		if err := runStep("Initializing git repository...", func() error {
			if _, err := runCmd(wd, opts.Verbose, "git", "init"); err != nil {
				return err
			}
			if _, err := runCmd(wd, opts.Verbose, "git", "add", "."); err != nil {
				return err
			}
			out, err := runCmd(wd, opts.Verbose, "git", "commit", "-m", "Initial commit")
			if err != nil && !opts.Verbose && len(out) > 0 {
				fmt.Println(string(out))
			}
			return err
		}, delay); err != nil {
			return err
		}
		_ = bar.Add(1)
	} else {
		color.Yellow("✔ Existing git repository detected. Skipping git init.")
	}

	_ = bar.Finish()
	fmt.Println()

	color.Green("\n🎉 Project ready!\n")
	color.Cyan("👉 Run: go run ./cmd\n")
	return nil
}

func runStep(message string, fn func() error, delay time.Duration) error {
	hippo := ui.NewHippoSpinner(os.Stdout)

	start := time.Now()
	hippo.Start(message)

	err := fn()

	minDuration := 600 * time.Millisecond
	elapsed := time.Since(start)

	if delay > 0 && elapsed < minDuration {
		time.Sleep(minDuration - elapsed)
	}

	hippo.Stop()

	if err != nil {
		color.Red("✖ %s", message)
		return err
	}

	color.Green("✔ %s", message)
	return nil
}

// runCmd: in non-verbose mode doesn't spam terminal, but collects output
func runCmd(dir string, verbose bool, name string, args ...string) ([]byte, error) {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir

	if verbose {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return nil, cmd.Run()
	}

	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	err := cmd.Run()

	out := buf.Bytes()
	if len(out) > 32*1024 {
		out = append(out[:32*1024], []byte("\n...output truncated...\n")...)
	}
	return out, err
}

func copyFromEmbed(src, dst, moduleName string) error {
	return fs.WalkDir(templates.FS, src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		if relPath == "." {
			return nil
		}

		targetPath := filepath.Join(dst, relPath)

		if d.IsDir() {
			return os.MkdirAll(targetPath, 0755)
		}

		// do not overwrite existing files
		if _, err := os.Stat(targetPath); err == nil {
			return nil
		}

		data, err := templates.FS.ReadFile(path)
		if err != nil {
			return err
		}

		content := strings.ReplaceAll(string(data), templateModuleName, moduleName)
		return os.WriteFile(targetPath, []byte(content), 0644)
	})
}

func hasGoMod(path string) bool {
	_, err := os.Stat(filepath.Join(path, "go.mod"))
	return err == nil
}

func hasGitRepo(path string) bool {
	_, err := os.Stat(filepath.Join(path, ".git"))
	return err == nil
}

func createEnvIfNotExists(wd string) error {
	envPath := filepath.Join(wd, ".env")
	if _, err := os.Stat(envPath); err == nil {
		return nil
	}

	content := `APPLICATION_HTTP_PORT=8080
APPLICATION_MODE=debug
TIMEZONE=Asia/Tashkent
LOG_LEVEL=debug
`
	return os.WriteFile(envPath, []byte(content), 0644)
}
