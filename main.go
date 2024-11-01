package gogit

import (
	"crypto/sha1"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

// Object interface for all git objects
type Object interface {
	Type() string
	Hash() string
	Serialize() []byte
}

// Blob object
type Blob struct {
	content []byte
}

func (b *Blob) Type() string {
	return "blob"
}

func (b *Blob) Hash() string {
	header := fmt.Sprintf("%s %d\x00", b.Type(), len(b.content))
	return fmt.Sprintf("%x", sha1.Sum([]byte(header+string(b.content))))
}

// Tree object
type TreeEntry struct {
	mode string
	name string
	hash string
}

type Tree struct {
	entries []TreeEntry
}

func (t *Tree) Type() string {
	return "tree"
}

// Commit object
type Commit struct {
	tree      string
	parent    string
	author    string
	committer string
	message   string
}

func (c *Commit) Type() string {
	return "commit"
}

type IndexEntry struct {
	ctime time.Time
	mtime time.Time
	dev   uint32
	ino   uint32
	mode  uint32
	uid   uint32
	gid   uint32
	size  uint32
	hash  string
	flags uint16
	path  string
}

type Index struct {
	entries map[string]*IndexEntry
}

func (repo *Repository) Add(paths []string) error {
	for _, path := range paths {
		// Read file content
		content, err := ioutil.ReadFile(filepath.Join(repo.workdir, path))
		if err != nil {
			return err
		}

		// Create blob object
		blob := &Blob{content: content}

		// Store blob in objects database
		if err := repo.storeObject(blob); err != nil {
			return err
		}

		// Update index
		stat, err := os.Stat(filepath.Join(repo.workdir, path))
		if err != nil {
			return err
		}

		entry := &IndexEntry{
			path:  path,
			hash:  blob.Hash(),
			size:  uint32(stat.Size()),
			mtime: stat.ModTime(),
			// ... set other metadata
		}

		repo.index.entries[path] = entry
	}

	return repo.writeIndex()
}

type Repository struct {
	workdir string
	gitdir  string
	index   *Index
}

func InitRepository(path string) (*Repository, error) {
	repo := &Repository{
		workdir: path,
		gitdir:  filepath.Join(path, ".git"),
	}

	// Create necessary directories
	dirs := []string{
		repo.gitdir,
		filepath.Join(repo.gitdir, "objects"),
		filepath.Join(repo.gitdir, "refs"),
		filepath.Join(repo.gitdir, "refs/heads"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, err
		}
	}

	// Initialize HEAD to point to master branch
	headFile := filepath.Join(repo.gitdir, "HEAD")
	if err := ioutil.WriteFile(headFile, []byte("ref: refs/heads/master\n"), 0644); err != nil {
		return nil, err
	}

	return repo, nil
}
