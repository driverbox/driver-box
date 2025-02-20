package main

import (
	"context"
	"fmt"
	"io"
	"math/bits"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/wailsapp/go-webview2/webviewloader"
	wails_runtime "github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx context.Context
}

func NewApp() *App {
	return &App{}
}

func (m *App) SetContext(ctx context.Context) {
	m.ctx = ctx
}

func (a *App) Cwd() (string, error) {
	if exePath, err := os.Executable(); err != nil {
		return "", err
	} else {
		return filepath.Dir(exePath), nil
	}
}

func (a *App) SelectFolder(relative bool) (string, error) {
	if path, err := wails_runtime.OpenDirectoryDialog(a.ctx, wails_runtime.OpenDialogOptions{}); err != nil || path == "" {
		return "", err
	} else if relative {
		if exePath, err := os.Executable(); err != nil {
			return "", err
		} else {
			return filepath.Rel(filepath.Dir(exePath), path)
		}
	} else {
		return path, nil
	}

}

func (a *App) SelectFile(relative bool) (string, error) {
	if path, err := wails_runtime.OpenFileDialog(a.ctx, wails_runtime.OpenDialogOptions{}); err != nil || path == "" {
		return "", err
	} else if relative {
		if exePath, err := os.Executable(); err != nil {
			return "", err
		} else {
			return filepath.Rel(filepath.Dir(exePath), path)
		}
	} else {
		return path, nil
	}

}

func (a App) PathExists(path string) bool {
	_, err := os.Stat(path)
	return err != nil
}

func (a App) ExecutableExists(path string) bool {
	_, err := exec.LookPath(path)
	return err == nil
}

func (a App) WebView2Version() (string, error) {
	return webviewloader.GetAvailableCoreWebView2BrowserVersionString(pathWV2)
}

func (a App) WebView2Path() string {
	return pathWV2
}

func (a App) AppConfigPath() string {
	return dirConf
}

func (a App) AppDriverPath() string {
	return dirDir
}

func (a App) AppVersion() string {
	return version.String()
}

func (a App) AppBinaryType() string {
	return fmt.Sprintf("%s%d", runtime.GOOS, bits.UintSize)
}

func (a App) Update(from string, to string, builtinWebview bool) error {
	file, err := os.CreateTemp("", "*.exe")
	if err != nil {
		return err
	}

	defer file.Close()

	response, err := http.Get(
		fmt.Sprintf("https://raw.githubusercontent.com/markmybytes/driver-box/updater/dist/updater.%s.exe", a.AppBinaryType()))
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if _, err := io.Copy(file, response.Body); err != nil {
		return err
	}

	file.Close()

	flags := []string{file.Name(), "-s", from, "-t", to, "-b", a.AppBinaryType()}
	if builtinWebview {
		flags = append(flags, "--webview")
	}

	process, err := os.StartProcess(file.Name(), flags, &os.ProcAttr{
		Dir:   ".",
		Env:   os.Environ(),
		Files: []*os.File{os.Stdin, os.Stdout, os.Stderr},
	})

	if err != nil {
		return err
	}

	if err := process.Release(); err != nil {
		return err
	}

	return nil
}
