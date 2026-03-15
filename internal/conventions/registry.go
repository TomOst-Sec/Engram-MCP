package conventions

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const communityRepoURL = "https://raw.githubusercontent.com/TomOst-Sec/engram-conventions/main/packs"

// PackRegistry manages community convention packs.
type PackRegistry struct {
	cacheDir string
}

// NewPackRegistry creates a new registry with packs cached in the given directory.
func NewPackRegistry(cacheDir string) *PackRegistry {
	return &PackRegistry{cacheDir: cacheDir}
}

// Install downloads a convention pack from the community repo.
func (r *PackRegistry) Install(ctx context.Context, name string) error {
	if err := os.MkdirAll(r.cacheDir, 0755); err != nil {
		return fmt.Errorf("create cache dir: %w", err)
	}

	url := fmt.Sprintf("%s/%s.json", communityRepoURL, name)
	client := &http.Client{Timeout: 10 * time.Second}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("download failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("convention pack '%s' not found. Browse available packs at https://github.com/TomOst-Sec/engram-conventions", name)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed: HTTP %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	// Validate the pack
	pack, err := ParsePack(data)
	if err != nil {
		return fmt.Errorf("invalid pack: %w", err)
	}

	// Save to cache
	dest := filepath.Join(r.cacheDir, name+".json")
	if err := os.WriteFile(dest, data, 0644); err != nil {
		return fmt.Errorf("save pack: %w", err)
	}

	fmt.Printf("Installed '%s' v%s (%d conventions)\n", pack.Name, pack.Version, len(pack.Conventions))
	return nil
}

// Remove deletes an installed convention pack.
func (r *PackRegistry) Remove(name string) error {
	path := filepath.Join(r.cacheDir, name+".json")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("pack '%s' is not installed", name)
	}
	return os.Remove(path)
}

// List returns all installed packs.
func (r *PackRegistry) List() ([]PackInfo, error) {
	entries, err := os.ReadDir(r.cacheDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var packs []PackInfo
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		data, err := os.ReadFile(filepath.Join(r.cacheDir, entry.Name()))
		if err != nil {
			continue
		}

		pack, err := ParsePack(data)
		if err != nil {
			continue
		}

		packs = append(packs, PackInfo{
			Name:        pack.Name,
			Version:     pack.Version,
			Description: pack.Description,
			Count:       len(pack.Conventions),
		})
	}

	return packs, nil
}

// LoadAllConventions loads conventions from all installed packs.
func (r *PackRegistry) LoadAllConventions() ([]Convention, error) {
	entries, err := os.ReadDir(r.cacheDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var all []Convention
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		data, err := os.ReadFile(filepath.Join(r.cacheDir, entry.Name()))
		if err != nil {
			continue
		}

		pack, err := ParsePack(data)
		if err != nil {
			continue
		}

		all = append(all, pack.Conventions...)
	}

	return all, nil
}
