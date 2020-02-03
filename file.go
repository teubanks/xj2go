package xj2go

import (
	"errors"
	"log"
	"os"
	"os/exec"
	"path"
	"regexp"
)

func checkFile(filename, pkg string) (string, error) {
	if ok, err := pathExists(pkg); !ok {
		os.Mkdir(pkg, 0755)
		if err != nil {
			return "", err
		}
	}

	filename = path.Base(filename)
	if filename[:1] == "." {
		return "", errors.New("File could not start with '.'")
	}

	filename = pkg + "/" + filename + ".go"
	if ok, _ := pathExists(filename); ok {
		if err := os.Remove(filename); err != nil {
			log.Fatal(err)
			return "", err
		}
	}

	return filename, nil
}

func writeStructToFile(filename, pkg string, strcts []strctMap) error {
	re := regexp.MustCompile("\\[|\\]")
	filename = re.ReplaceAllString(filename, "")

	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	defer file.Close()
	if err != nil {
		log.Fatal(err)
		return err
	}

	err = writeStruct(file, pkg, strcts)
	if err != nil {
		log.Fatal(err)
		return err
	}

	ft := exec.Command("go", "fmt", filename)
	if err := ft.Run(); err != nil {
		log.Fatal(err)
		return err
	}

	vt := exec.Command("go", "vet", filename)
	if err := vt.Run(); err != nil {
		log.Fatal(err)
		return err
	}

	return nil
}
