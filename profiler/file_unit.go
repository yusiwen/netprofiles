package profiler

import (
	"log"
	"os"
	"path/filepath"

	"github.com/yusiwen/netprofiles/utils"
)

type PostLoadFunc func() error
type EmptyHandlerFunc func() error

type File struct {
	Path          string `json:"path"`
	RootPrivilege bool   `json:"root-privilege"`
}

type FileUnit struct {
	Name         string           `json:"name"`
	Files        []File           `json:"files"`
	PostLoad     PostLoadFunc     `json:"-"`
	EmptyHandler EmptyHandlerFunc `json:"-"`
}

func (fp *FileUnit) Save(profile, location string) error {
	for _, f := range fp.Files {
		if !utils.Exists(f.Path) {
			log.Printf("warning: '%s' not found", f.Path)
			if fp.EmptyHandler != nil {
				err := fp.EmptyHandler()
				if err != nil {
					return err
				}
			}
			continue
		}
		dstPath := filepath.Join(location, profile, fp.Name)
		err := os.MkdirAll(dstPath, os.ModePerm)
		if err != nil {
			return err
		}
		dst := filepath.Join(dstPath, filepath.Base(f.Path))
		_, err = utils.Copy(f.Path, dst)
		if err != nil {
			return err
		}
		log.Printf("save '%s' to '%s'\n", f.Path, dst)
	}
	return nil
}

func (fp *FileUnit) Load(profile, location string) error {
	for _, f := range fp.Files {
		srcPath := filepath.Join(location, profile, fp.Name)
		src := filepath.Join(srcPath, filepath.Base(f.Path))
		log.Printf("loading '%s' ...", src)
		if !utils.Exists(src) {
			log.Printf("loading '%s' ... skip", src)
			continue
		}
		var err error
		if f.RootPrivilege {
			err = utils.CopySudo(src, f.Path)
		} else {
			_, err = utils.Copy(src, f.Path)
		}
		if err != nil {
			return err
		}
		log.Printf("loading '%s' to '%s'\n", src, f.Path)
	}
	if fp.PostLoad != nil {
		err := fp.PostLoad()
		if err != nil {
			return err
		}
	}
	return nil
}
