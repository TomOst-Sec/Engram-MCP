package testdata

import (
	"fmt"
	"os"
	"path/filepath"
)

// GenerateRepo creates a synthetic repository with the specified number of files.
// Files are split: 40% Go, 30% Python, 30% TypeScript.
func GenerateRepo(dir string, fileCount int) error {
	goCount := fileCount * 40 / 100
	pyCount := fileCount * 30 / 100
	tsCount := fileCount - goCount - pyCount

	for i := range goCount {
		path := filepath.Join(dir, fmt.Sprintf("pkg/module%d/handler%d.go", i/10, i))
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return err
		}
		if err := os.WriteFile(path, goFile(i), 0644); err != nil {
			return err
		}
	}

	for i := range pyCount {
		path := filepath.Join(dir, fmt.Sprintf("lib/module%d/service%d.py", i/10, i))
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return err
		}
		if err := os.WriteFile(path, pyFile(i), 0644); err != nil {
			return err
		}
	}

	for i := range tsCount {
		path := filepath.Join(dir, fmt.Sprintf("src/components/Component%d.ts", i))
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return err
		}
		if err := os.WriteFile(path, tsFile(i), 0644); err != nil {
			return err
		}
	}

	return nil
}

func goFile(n int) []byte {
	return []byte(fmt.Sprintf(`package module%d

import (
	"context"
	"fmt"
	"net/http"
)

// Handler%d processes incoming requests for module %d.
// It validates input, applies business logic, and returns a response.
func Handler%d(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	if r.Method != "POST" {
		return fmt.Errorf("method not allowed: %%s", r.Method)
	}
	return nil
}

// Process%d applies the core business logic for this module.
func Process%d(input string) (string, error) {
	if input == "" {
		return "", fmt.Errorf("empty input")
	}
	return fmt.Sprintf("processed: %%s", input), nil
}

type Service%d struct {
	Name    string
	Version int
	Active  bool
}

func NewService%d(name string) *Service%d {
	return &Service%d{Name: name, Version: 1, Active: true}
}

func (s *Service%d) Start() error {
	if !s.Active {
		return fmt.Errorf("service %%s is inactive", s.Name)
	}
	return nil
}

func (s *Service%d) Stop() error {
	s.Active = false
	return nil
}
`, n/10, n, n, n, n, n, n, n, n, n, n, n))
}

func pyFile(n int) []byte {
	return []byte(fmt.Sprintf(`"""Module %d service implementation."""

from typing import Optional, List

class Service%d:
    """Service%d handles business logic for module %d."""

    def __init__(self, name: str, version: int = 1):
        self.name = name
        self.version = version
        self._active = True

    def process(self, data: str) -> str:
        """Process input data and return result."""
        if not data:
            raise ValueError("empty input")
        return f"processed: {data}"

    def validate(self, items: List[str]) -> bool:
        """Validate a list of items."""
        return all(len(item) > 0 for item in items)

    @property
    def is_active(self) -> bool:
        return self._active

def create_service%d(name: str) -> Service%d:
    """Factory function for Service%d."""
    return Service%d(name)

def helper_function%d(x: int, y: int) -> int:
    """Compute sum of two integers."""
    return x + y
`, n, n, n, n, n, n, n, n, n))
}

func tsFile(n int) []byte {
	return []byte(fmt.Sprintf(`/**
 * Component%d - UI component for feature %d
 */

interface Props%d {
  title: string;
  count: number;
  active: boolean;
}

export function Component%d(props: Props%d): string {
  const { title, count, active } = props;
  if (!active) {
    return "";
  }
  return title + " (" + count + ")";
}

export function useFeature%d(id: number): Props%d {
  return {
    title: "Feature " + id,
    count: 0,
    active: true,
  };
}

export class Manager%d {
  private items: string[] = [];

  add(item: string): void {
    this.items.push(item);
  }

  getAll(): string[] {
    return [...this.items];
  }

  count(): number {
    return this.items.length;
  }
}
`, n, n, n, n, n, n, n, n))
}
