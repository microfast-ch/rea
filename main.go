package main

import (
	"archive/zip"
	"io/ioutil"
	"log"
)

func main() {
	r, err := zip.OpenReader("Basic1.ott")
	if err != nil {
		log.Fatal(err)
	}
	defer r.Close()

	for _, f := range r.File {
		log.Println(f.Name)
		log.Println(f.Comment)
		if f.Name != "mimetype" {
			continue
		}

		fd, err := f.Open()
		if err != nil {
			log.Fatal(err)
		}

		b, err := ioutil.ReadAll(fd)
		if err != nil {
			log.Fatal(err)
		}
		fd.Close()

		log.Printf("metadata: %s", string(b))
		if string(b) != "application/vnd.oasis.opendocument.text-template" {
			log.Println("wrong mimetype, abort")
			return
		}

	}
}
