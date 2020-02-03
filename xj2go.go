package xj2go

import (
	"bytes"
	"encoding/xml"
	"io"
	"log"
	"sort"
)

// XJ define xj2go struct
type XJ struct {
	// xml or json file
	Filepath string
	// the pkg name for struct
	Pkgname string
	// the root name for json bytes
	Rootname string
}

// New return a xj2go instance
func New(filepath, pkgname, rootname string) *XJ {
	return &XJ{
		Filepath: filepath,
		Pkgname:  pkgname,
		Rootname: rootname,
	}
}

// XMLToGo convert xml to go struct, then write this struct to a go file
func (xj *XJ) XMLToGo() error {
	filename, err := checkFile(xj.Filepath, xj.Pkgname)
	if err != nil {
		log.Fatal(err)
		return err
	}

	nodes, err := xmlToLeafNodes(xj.Filepath)
	if err != nil {
		log.Fatal(err)
		return err
	}
	strcts := leafNodesToStrcts("xml", nodes)

	return writeStructToFile(filename, xj.Pkgname, strcts)
}

// XMLBytesToGo convert xml bytes to struct, then the struct will be writed to ./{pkg}/{filename}.go
func XMLBytesToGo(filename, pkgname string, b *[]byte) error {
	filename, err := checkFile(filename, pkgname)
	if err != nil {
		log.Fatal(err)
		return err
	}

	r := bytes.NewReader(*b)
	m, err := decodeXML(xml.NewDecoder(r), "", nil)
	if err != nil {
		log.Fatal(err)
		return err
	}

	nodes, err := leafNodes(m)
	if err != nil {
		log.Fatal(err)
		return err
	}
	strcts := leafNodesToStrcts("xml", nodes)

	return writeStructToFile(filename, pkgname, strcts)
}

// JSONToGo convert json to go struct, then write this struct to a go file
func (xj *XJ) JSONToGo() error {
	filename, err := checkFile(xj.Filepath, xj.Pkgname)
	if err != nil {
		log.Fatal(err)
		return err
	}

	nodes, err := jsonToLeafNodes(xj.Rootname, xj.Filepath)
	if err != nil {
		log.Fatal(err)
		return err
	}
	strcts := leafNodesToStrcts("json", nodes)

	return writeStructToFile(filename, xj.Pkgname, strcts)
}

// JSONBytesToGoFile convert json bytes to struct, then the struct will be writed to ./{pkg}/{filename}.go
func JSONBytesToGoFile(filename, pkgname, rootname string, b *[]byte) error {
	filename, err := checkFile(filename, pkgname)
	if err != nil {
		log.Fatal(err)
		return err
	}

	m, err := jsonBytesToMap(rootname, b)
	if err != nil {
		log.Fatal(err)
		return err
	}

	ns, err := leafNodes(m)
	if err != nil {
		log.Fatal(err)
		return err
	}

	nodes, err := reLeafNodes(ns, rootname)
	if err != nil {
		log.Fatal(err)
		return err
	}

	strcts := leafNodesToStrcts("json", nodes)

	return writeStructToFile(filename, pkgname, strcts)
}

func JSONBytesToGo(pkg, rootname string, data *[]byte) ([]byte, error) {
	m, err := jsonBytesToMap(rootname, data)
	if err != nil {
		log.Fatal(err)
		return []byte(""), err
	}

	ns, err := leafNodes(m)
	if err != nil {
		log.Fatal(err)
		return []byte(""), err
	}

	nodes, err := reLeafNodes(ns, rootname)
	if err != nil {
		log.Fatal(err)
		return []byte(""), err
	}

	strcts := leafNodesToStrcts("json", nodes)
	w := &bytes.Buffer{}
	err = writeStruct(w, pkg, strcts)
	if err != nil {
		log.Fatal(err)
		return []byte(""), err
	}
	return w.Bytes(), nil
}

func writeStruct(w io.Writer, pkg string, strcts []strctMap) error {
	pkgLines := make(map[string]string)
	strctLines := []string{}

	var roots []string
	strctsMap := make(map[string]strctMap)

	for _, strct := range strcts {
		for root := range strct {
			roots = append(roots, root)
			strctsMap[root] = strct
		}
	}

	sort.Strings(roots)

	for _, root := range roots {
		strct := strctsMap[root]
		for r, sns := range strct {
			sort.Sort(byName(sns))
			strctLines = append(strctLines, "type "+toProperCase(r)+" struct {\n")
			for i := 0; i < len(sns); i++ {
				if sns[i].Type == "time.Time" {
					pkgLines["time.Time"] = "import \"time\"\n"
				}
				strctLines = append(strctLines, "\t"+toProperCase(sns[i].Name)+"\t"+sns[i].Type+"\t"+sns[i].Tag+"\n")
			}
			strctLines = append(strctLines, "}\n")
		}
	}

	strctLines = append(strctLines, "\n")

	if pkg != "" {
		_, err := w.Write([]byte("package " + pkg + "\n\n"))
		if err != nil {
			return err
		}
		for _, pl := range pkgLines {
			_, err = w.Write([]byte(pl))
			if err != nil {
				return err
			}
		}
	}
	for _, sl := range strctLines {
		_, err := w.Write([]byte(sl))
		if err != nil {
			return err
		}
	}

	return nil
}
