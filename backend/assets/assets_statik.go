package assets

import (
	"errors"
	"io"
	"os"
	"syscall"
	"time"

	"net/http"
	"path/filepath"

	"github.com/dkumor/statik/fs"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

// The cached builtin Assets
var builtinAssets afero.Fs

type AferoHttpFile struct {
	http.File
}

func (a AferoHttpFile) Name() string {
	s, err := a.File.Stat()
	if err != nil {
		return "" // No errors allowed
	}
	return s.Name()
}

func (a AferoHttpFile) WriteAt(p []byte, off int64) (int, error) {
	return 0, syscall.EPERM
}

func (a AferoHttpFile) ReadAt(p []byte, off int64) (int, error) {
	if _, err := a.File.Seek(off, io.SeekStart); err != nil {
		return 0, err
	}

	return a.File.Read(p)
}

func (a AferoHttpFile) Readdirnames(n int) ([]string, error) {
	dirs, err := a.File.Readdir(n)
	if err != nil {
		return nil, err
	}

	out := make([]string, len(dirs))
	for d := range dirs {
		out[d] = dirs[d].Name()
	}
	return out, nil
}

func (a AferoHttpFile) Sync() error {
	return nil
}

func (a AferoHttpFile) Truncate(size int64) error {
	return syscall.EPERM
}

func (a AferoHttpFile) WriteString(s string) (int, error) {
	return 0, syscall.EPERM
}

func (a AferoHttpFile) Write(n []byte) (int, error) {
	return 0, syscall.EPERM
}

type AferoReverseHttpFs struct {
	http.FileSystem
}

func NewAferoReverseHttpFs(fs http.FileSystem) AferoReverseHttpFs {
	return AferoReverseHttpFs{fs}
}

func (fs AferoReverseHttpFs) Mkdir(n string, p os.FileMode) error {
	return syscall.EPERM
}

func (r AferoReverseHttpFs) MkdirAll(n string, p os.FileMode) error {
	return syscall.EPERM
}

func (fs AferoReverseHttpFs) Create(n string) (afero.File, error) {
	return nil, syscall.EPERM
}

func (fs AferoReverseHttpFs) ReadDir(name string) ([]os.FileInfo, error) {
	f, err := fs.FileSystem.Open(name)
	if err != nil {
		return nil, err
	}
	return f.Readdir(0)
}

func (fs AferoReverseHttpFs) Chtimes(n string, a, m time.Time) error {
	return syscall.EPERM
}

func (fs AferoReverseHttpFs) Chmod(n string, m os.FileMode) error {
	return syscall.EPERM
}

func (fs AferoReverseHttpFs) Name() string {
	return "AferoPackr"
}

func (fs AferoReverseHttpFs) Stat(name string) (os.FileInfo, error) {
	f, err := fs.FileSystem.Open(name)
	if err != nil {
		return nil, err
	}
	return f.Stat()
}

func (fs AferoReverseHttpFs) LstatIfPossible(name string) (os.FileInfo, bool, error) {
	fi, err := fs.Stat(name)
	return fi, false, err
}

func (fs AferoReverseHttpFs) Rename(o, n string) error {
	return syscall.EPERM
}

func (fs AferoReverseHttpFs) RemoveAll(p string) error {
	return syscall.EPERM
}

func (fs AferoReverseHttpFs) Remove(n string) error {
	return syscall.EPERM
}

func (fs AferoReverseHttpFs) OpenFile(name string, flag int, perm os.FileMode) (afero.File, error) {
	if flag&(os.O_WRONLY|syscall.O_RDWR|os.O_APPEND|os.O_CREATE|os.O_TRUNC) != 0 {
		return nil, syscall.EPERM
	}

	return fs.Open(name)
}

func (fs AferoReverseHttpFs) Open(n string) (afero.File, error) {
	f, err := fs.FileSystem.Open(n)
	return AferoHttpFile{f}, err
}

func BuiltinAssets() afero.Fs {
	if builtinAssets == nil {
		statikFS, err := fs.New()
		if err != nil {
			// Try to find an assets folder in the ancestors
			cwd, err := os.Getwd()
			if err != nil {
				panic(err)
			}
			assetPath := filepath.Join(cwd, "assets")
			_, err = os.Stat(filepath.Join(assetPath, "heedy.conf"))
			for os.IsNotExist(err) {
				cwdnew := filepath.Dir(cwd)
				if cwdnew == cwd {
					panic(errors.New("Could not find assets folder"))
				}
				cwd = cwdnew
				assetPath = filepath.Join(cwd, "assets")
				_, err = os.Stat(filepath.Join(assetPath, "heedy.conf"))
			}

			log.Warnf("Debug mode: using assets from %s", assetPath)

			builtinAssets = afero.NewBasePathFs(afero.NewOsFs(), assetPath)
		} else {
			builtinAssets = NewAferoReverseHttpFs(statikFS)
		}

	}

	return builtinAssets
}
