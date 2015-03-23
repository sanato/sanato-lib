package storage

import (
	"fmt"
	"io"
	"mime"
	"os"
	"path/filepath"
)

type MetaData struct {
	Id           string      `json:"id"`
	Path         string      `json:"path"`
	Size         int64       `json:"size"`
	IsCol        bool        `json:"isCol"`
	MimeType     string      `json:"mimeType"`
	Checksum     string      `json:"checksum"`
	ChecksumType string      `json:"checksumType"`
	Modified     int64       `json:"modified"`
	ETag         string      `json:"etag"`
	Children     []*MetaData `json:"children"`
	Extra        interface{} `json:"extra"` // maybe to save xattrs or custom user data
}

func NewStorageProvider(rootDataDir string) (*StorageProvider, error) {
	return &StorageProvider{rootDataDir}, nil
}

type StorageProvider struct {
	rootDataDir string
}

func (sp *StorageProvider) PutFile(path string, r io.Reader, size int64) error {
	fullpath := filepath.Join(sp.rootDataDir, "/", path)

	fd, err := os.Create(fullpath)
	defer fd.Close()
	if err != nil {
		return convertError(err)
	}

	/* TEST WHY THIS DONT WORK MAYBE BECAUSE IS CHUNK UPLOAD AND WE DONT HAVE CONTENT-LENGTH?
	if _, err := io.CopyN(fd, r, size); err != nil {
		return err
	}
	*/

	_, err = io.Copy(fd, r)
	return convertError(err)

	/*if checksumType == "md5" {
		sum := md5.Sum(data)
		if fmt.Sprintf("%x", sum) != checksum {
			return errors.New("invalid checksum")
		}
	}
	*/
}
func (sp *StorageProvider) Stat(path string, children bool) (*MetaData, error) {
	fullpath := filepath.Join(sp.rootDataDir, "/", path)

	finfo, err := os.Stat(fullpath)
	if err != nil {
		return nil, convertError(err)
	}

	mimeType := mime.TypeByExtension(filepath.Ext(path))
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}
	if finfo.IsDir() {
		mimeType = "inode/directory"
	}
	meta := MetaData{
		Id:       path,
		Path:     path,
		Size:     finfo.Size(),
		IsCol:    finfo.IsDir(),
		Modified: finfo.ModTime().Unix(),
		ETag:     fmt.Sprintf("\"%d\"", finfo.ModTime().Unix()),
		MimeType: mimeType,
	}

	if meta.IsCol == false {
		return &meta, nil
	}
	if children == false {
		return &meta, nil
	}

	fd, err := os.Open(fullpath)
	defer fd.Close()
	if err != nil {
		return nil, convertError(err)
	}

	finfos, err := fd.Readdir(0)
	if err != nil {
		return nil, convertError(err)
	}

	meta.Children = make([]*MetaData, len(finfos))
	for i, f := range finfos {
		childPath := filepath.Join(path, "/", f.Name())
		mimeType := mime.TypeByExtension(filepath.Ext(childPath))
		if mimeType == "" {
			mimeType = "application/octet-stream"
		}
		if f.IsDir() {
			mimeType = "inode/directory"
		}
		m := MetaData{
			Id:       childPath,
			Path:     childPath,
			Size:     f.Size(),
			IsCol:    f.IsDir(),
			Modified: f.ModTime().Unix(),
			ETag:     fmt.Sprintf("\"%d\"", f.ModTime().Unix()),
			MimeType: mimeType,
		}
		meta.Children[i] = &m
	}

	return &meta, nil
}

func (sp *StorageProvider) GetFile(path string) (io.Reader, error) {
	fullpath := filepath.Join(sp.rootDataDir, "/", path)
	file, err := os.Open(fullpath)
	if err != nil {
		return nil, convertError(err)
	}
	return file, nil
}

//  Delete removed the resource identified by the id param or it returns error
//  in case of failure.
func (sp *StorageProvider) Remove(path string, recursive bool) error {
	fullpath := filepath.Join(sp.rootDataDir, "/", path)
	if !recursive {
		return convertError(os.Remove(fullpath))
	}
	return convertError(os.RemoveAll(fullpath))
}

func (sp *StorageProvider) CreateCol(path string, recursive bool) error {
	fullpath := filepath.Join(sp.rootDataDir, "/", path)
	if recursive == false {
		return convertError(os.Mkdir(fullpath, 0666))
	}
	return convertError(os.MkdirAll(fullpath, 0666))
}

func (sp *StorageProvider) Copy(from, to string) error {
	fromfullpath := filepath.Join(sp.rootDataDir, "/", from)
	tofullpath := filepath.Join(sp.rootDataDir, "/", to)
	src, err := os.Open(fromfullpath)
	defer src.Close()
	if err != nil {
		return err
	}
	dst, err := os.Create(tofullpath)
	defer dst.Close()
	if err != nil {
		return err
	}
	_, err = io.Copy(dst, src)
	return err
}
func (sp *StorageProvider) Rename(from, to string) error {
	fromfullpath := filepath.Join(sp.rootDataDir, "/", from)
	tofullpath := filepath.Join(sp.rootDataDir, "/", to)
	return convertError(os.Rename(fromfullpath, tofullpath))
}

type ExistError struct {
	Op   string
	Path string
	Err  error
}

func (e *ExistError) Error() string { return e.Op + " " + e.Path + ": " + e.Err.Error() }

type NotExistError struct {
	Op   string
	Path string
	Err  error
}

func (e *NotExistError) Error() string { return e.Op + " " + e.Path + ": " + e.Err.Error() }

func IsExistError(err error) bool {
	if err == nil {
		return false
	}
	_, ok := err.(*ExistError)
	if !ok {
		return false
	}
	return true
}
func IsNotExistError(err error) bool {
	if err == nil {
		return false
	}
	_, ok := err.(*NotExistError)
	if !ok {
		return false
	}
	return true
}

func checkExistError(err error) error {
	if !os.IsExist(err) {
		return nil
	}
	te, ok := err.(*os.PathError)
	if !ok {
		return err
	}
	return &ExistError{te.Op, te.Path, te.Err}
}
func checkNotExistError(err error) error {
	if !os.IsNotExist(err) {
		return nil
	}
	te, ok := err.(*os.PathError)
	if !ok {
		return err
	}
	return &NotExistError{te.Op, te.Path, te.Err}
}
func convertError(err error) error {
	if err == nil {
		return nil
	}
	if te := checkExistError(err); te != nil {
		return te
	}
	if te := checkNotExistError(err); te != nil {
		return te
	}
	return err
}
