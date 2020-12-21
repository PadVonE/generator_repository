package usecase

import (
	log "github.com/sirupsen/logrus"
	"os"
	"path/filepath"
)

func RemoveContents(dir string) error {
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer d.Close()
	names, err := d.Readdirnames(-1)
	if err != nil {
		return err
	}
	for _, name := range names {
		err = os.RemoveAll(filepath.Join(dir, name))
		if err != nil {
			return err
		}
	}
	return nil
}

func FileSave(saveFilePath,code string) (err error) {
	saveFile, err := os.Create(saveFilePath)

	if err != nil {
		log.Error("Unable to create file:", err)
		return
	}

	defer saveFile.Close()
	saveFile.WriteString(code)

	return
}