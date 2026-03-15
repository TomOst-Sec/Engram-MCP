package indexer

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/TomOst-Sec/colony-project/internal/config"
	"github.com/TomOst-Sec/colony-project/internal/parser"
	"github.com/fsnotify/fsnotify"
)

// Watcher monitors the repository for file changes and re-indexes automatically.
type Watcher struct {
	indexer  *Indexer
	repoRoot string
	config   *config.Config
	done     chan struct{}
	watcher  *fsnotify.Watcher

	mu      sync.Mutex
	timers  map[string]*time.Timer
}

// NewWatcher creates a new file system watcher.
func NewWatcher(indexer *Indexer, repoRoot string, cfg *config.Config) *Watcher {
	return &Watcher{
		indexer:  indexer,
		repoRoot: repoRoot,
		config:   cfg,
		done:     make(chan struct{}),
		timers:   make(map[string]*time.Timer),
	}
}

// Start begins watching the repository for file changes.
// Blocks until ctx is cancelled or Stop() is called.
func (w *Watcher) Start(ctx context.Context) error {
	var err error
	w.watcher, err = fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("creating watcher: %w", err)
	}
	defer w.watcher.Close()

	// Walk repo and add directories
	if err := w.addDirectories(); err != nil {
		return fmt.Errorf("adding directories: %w", err)
	}

	log.Printf("Watching %s for changes...", w.repoRoot)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-w.done:
			return nil
		case event, ok := <-w.watcher.Events:
			if !ok {
				return nil
			}
			w.handleEvent(ctx, event)
		case err, ok := <-w.watcher.Errors:
			if !ok {
				return nil
			}
			log.Printf("Watcher error: %v", err)
		}
	}
}

// Stop signals the watcher to stop.
func (w *Watcher) Stop() {
	close(w.done)
}

func (w *Watcher) handleEvent(ctx context.Context, event fsnotify.Event) {
	absPath := event.Name

	// Check if this is a directory event (new dir created)
	if event.Has(fsnotify.Create) {
		if info, err := os.Stat(absPath); err == nil && info.IsDir() {
			if !w.shouldIgnoreDir(filepath.Base(absPath)) {
				if err := w.watcher.Add(absPath); err != nil {
					log.Printf("Warning: could not watch new dir %s: %v", absPath, err)
				}
			}
			return
		}
	}

	relPath, err := filepath.Rel(w.repoRoot, absPath)
	if err != nil {
		return
	}

	// Check ignore patterns on path components
	if w.shouldIgnorePath(relPath) {
		return
	}

	// Check if file has a registered parser
	if _, ok := w.indexer.registry.ParserFor(relPath); !ok {
		return
	}

	if event.Has(fsnotify.Remove) || event.Has(fsnotify.Rename) {
		// File removed — delete symbols
		w.debounce(relPath, func() {
			if err := parser.DeleteFileSymbols(w.indexer.store, relPath); err != nil {
				log.Printf("Error deleting symbols for %s: %v", relPath, err)
			} else {
				log.Printf("Removed symbols for %s", relPath)
			}
		})
		return
	}

	if event.Has(fsnotify.Create) || event.Has(fsnotify.Write) {
		w.debounce(relPath, func() {
			start := time.Now()
			source, err := os.ReadFile(absPath)
			if err != nil {
				log.Printf("Error reading %s: %v", relPath, err)
				return
			}

			// Skip large files
			if int64(len(source)) > w.config.MaxFileSize {
				return
			}

			count, err := w.indexer.IndexFile(ctx, relPath, source)
			if err != nil {
				log.Printf("Error indexing %s: %v", relPath, err)
				return
			}
			elapsed := time.Since(start)
			log.Printf("Re-indexed %s (%d symbols, %dms)", relPath, count, elapsed.Milliseconds())
		})
	}
}

func (w *Watcher) debounce(path string, fn func()) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if t, ok := w.timers[path]; ok {
		t.Stop()
	}
	w.timers[path] = time.AfterFunc(100*time.Millisecond, func() {
		fn()
		w.mu.Lock()
		delete(w.timers, path)
		w.mu.Unlock()
	})
}

func (w *Watcher) addDirectories() error {
	return filepath.WalkDir(w.repoRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.IsDir() {
			return nil
		}
		name := d.Name()
		if w.shouldIgnoreDir(name) {
			return filepath.SkipDir
		}
		if err := w.watcher.Add(path); err != nil {
			log.Printf("Warning: could not watch %s: %v", path, err)
		}
		return nil
	})
}

func (w *Watcher) shouldIgnoreDir(name string) bool {
	if strings.HasPrefix(name, ".") && name != "." {
		return true
	}
	for _, pattern := range w.config.IgnorePatterns {
		trimmed := strings.TrimSuffix(pattern, "/")
		if name == trimmed {
			return true
		}
	}
	return false
}

func (w *Watcher) shouldIgnorePath(relPath string) bool {
	parts := strings.Split(relPath, string(filepath.Separator))
	for _, part := range parts {
		if w.shouldIgnoreDir(part) {
			return true
		}
	}
	return false
}
