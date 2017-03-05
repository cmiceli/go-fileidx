package walker

import (
	"crypto/sha1"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
)

import (
	"github.com/golang/glog"
)

type FileInfo struct {
	Hash     string
	Filename string
	Fileobj  os.FileInfo
}

func hash_file_sha1(filePath string) (string, error) {
	var returnSHA1String string

	file, err := os.Open(filePath)
	if err != nil {
		return returnSHA1String, err
	}

	defer file.Close()

	hash := sha1.New()

	if _, err := io.Copy(hash, file); err != nil {
		return returnSHA1String, err
	}
	hashInBytes := hash.Sum(nil)[:20]
	returnSHA1String = hex.EncodeToString(hashInBytes)
	return returnSHA1String, nil
}

type Walker struct {
	fileChan chan<- FileInfo
	errChan  chan<- error
}

func NewWalker(fileChan chan<- FileInfo, errChan chan<- error) *Walker {
	return &Walker{fileChan: fileChan, errChan: errChan}
}

func (w *Walker) walkFunc(path string, fileinfo os.FileInfo, err error) error {
	glog.Infof("Hit the file: %s", path)
	if err != nil {
		glog.Errorf("Ran into an error: %s", path)
		w.errChan <- err
		return nil
	}
	if fileinfo.IsDir() {
		return nil
	}
	glog.Infof("Hashing: %s", path)
	hash, err := hash_file_sha1(path)
	if err != nil {
		glog.Errorf("Error hashing: %s", path)
		w.errChan <- err
		return nil
	}

	glog.Infof("Pushing file into channel: %s", path)
	w.fileChan <- FileInfo{Hash: hash, Filename: path, Fileobj: fileinfo}
	return nil
}

func (w *Walker) Walk(srcDir string) {
	filepath.Walk(srcDir, w.walkFunc)
	close(w.fileChan)
	close(w.errChan)
}
