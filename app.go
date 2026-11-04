package main

import (
	"context"
	"errors"
	"os"
	"path/filepath"
)

// App holds the backend state that frontend bindings can call.
type App struct {
	ctx context.Context
}

// NewApp creates the backend application instance.
func NewApp() *App {
	return &App{}
}

// startup is called by Wails when the window has been created.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	migrateLegacyFiles()
}

// exeDir returns the directory that contains the running executable.
func exeDir() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	if resolved, resolveErr := filepath.EvalSymlinks(exe); resolveErr == nil {
		exe = resolved
	}
	return filepath.Dir(exe), nil
}

// dataDir is where the account lives: the per-user configuration directory on
// Windows (e.g. C:\Users\<name>\AppData\Roaming\TextCrypto). Keeping it here
// means the account stays on the machine even when the exe is moved, replaced
// or rebuilt. If that directory is unavailable, fall back to the exe folder.
func dataDir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil || base == "" {
		return exeDir()
	}

	dir := filepath.Join(base, "TextCrypto")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return exeDir()
	}
	return dir, nil
}

func accountFilePath() (string, error) {
	dir, err := dataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "account.json"), nil
}

// migrationMarkerPath records that accounts stored by older builds were already
// imported, so the legacy IndexedDB vault is never read twice.
func migrationMarkerPath() (string, error) {
	dir, err := dataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "account.migrated"), nil
}

// migrateLegacyFiles moves account files created by earlier builds (which kept
// them next to the exe) into the user data directory. It never overwrites an
// existing file and never deletes a source it failed to copy.
func migrateLegacyFiles() {
	target, err := dataDir()
	if err != nil {
		return
	}
	source, err := exeDir()
	if err != nil || source == target {
		return
	}

	for _, name := range []string{"account.json", "account.migrated"} {
		src := filepath.Join(source, name)
		dst := filepath.Join(target, name)

		if _, statErr := os.Stat(dst); statErr == nil {
			continue
		}

		data, readErr := os.ReadFile(src)
		if readErr != nil {
			continue
		}
		if writeErr := os.WriteFile(dst, data, 0o600); writeErr != nil {
			continue
		}
		os.Remove(src)
	}
}

// LoadAccount returns the raw contents of account.json, or an empty string
// when the file does not exist yet.
func (a *App) LoadAccount() (string, error) {
	migrateLegacyFiles()

	path, err := accountFilePath()
	if err != nil {
		return "", err
	}

	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// SaveAccount writes the encrypted envelope into the user data directory.
// The write goes to a temporary file first and is then renamed, so a crash
// can never leave a half-written account file behind.
func (a *App) SaveAccount(content string) error {
	path, err := accountFilePath()
	if err != nil {
		return err
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}

	tmp, err := os.CreateTemp(dir, ".account-*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)

	if _, err := tmp.WriteString(content); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Chmod(tmpName, 0o600); err != nil {
		// Permission bits are advisory on Windows; ignore failures.
		_ = err
	}
	return os.Rename(tmpName, path)
}

// AccountPath exposes the resolved account file location.
func (a *App) AccountPath() (string, error) {
	return accountFilePath()
}

// IsMigrated reports whether account.json is already the single source of truth.
func (a *App) IsMigrated() (bool, error) {
	path, err := migrationMarkerPath()
	if err != nil {
		return false, err
	}

	_, statErr := os.Stat(path)
	if statErr == nil {
		return true, nil
	}
	if errors.Is(statErr, os.ErrNotExist) {
		return false, nil
	}
	return false, statErr
}

// MarkMigrated records that the account no longer depends on IndexedDB.
func (a *App) MarkMigrated() error {
	path, err := migrationMarkerPath()
	if err != nil {
		return err
	}
	return os.WriteFile(path, []byte("account data lives in account.json\n"), 0o600)
}
