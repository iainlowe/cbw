package main

import (
	"archive/zip"
	"bufio"
	"flag"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/djimenez/iconv-go"

	"github.com/dustin/go-humanize"
)

var (
	zipfilelist = flag.String("f", "data/ziplist", "File with list of ZIP files to download (one URL per line)")
	tmpdir      = flag.String("tmpdir", "zips", "Temporary directory for storing ZIP files")
	destdir     = flag.String("d", "mp3", "Destination dir for MP3 files")
)

func zipfiles() (zfs chan string) {
	zfs = make(chan string, 1)

	var (
		err  error
		file *os.File
	)

	go func() {
		defer close(zfs)

		if file, err = os.Open(*zipfilelist); err != nil {
			panic(err)
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			if line := strings.TrimSpace(scanner.Text()); line != "" {
				zfs <- line
			}
		}

		if err := scanner.Err(); err != nil {
			panic(err)
		}
	}()

	return zfs
}

func extract(url string) {
	zipname := filepath.Base(url)
	fpath := filepath.Join(*tmpdir, zipname)

	var (
		zr      *zip.ReadCloser // ZIP archive reader
		zf      *zip.File       // zipped file in archive
		resp    *http.Response  // HTTP response for ZIP file request
		zipfile *os.File        // written file from HTTP download

		err error
	)

	if _, err := os.Stat(fpath); err == nil {
		log.Println("Skipping download of", url)
	} else if resp, err = http.Get(url); err != nil {
		panic(err)
	} else if zipfile, err = os.Create(fpath); err != nil {
		panic(err)
	} else {
		log.Println("Downloading", url)
		io.Copy(zipfile, resp.Body)
		resp.Body.Close()
		zipfile.Close()
	}

	if zr, err = zip.OpenReader(fpath); err != nil {
		panic(err)
	}
	defer zr.Close()

	log.Println("Extracting files from", zipname, "to", *destdir)

	for _, zf = range zr.File {
		if zf.FileInfo().IsDir() {
			continue
		}

		extractFile(zf)
	}
}

func extractFile(zf *zip.File) {
	var (
		zfbody  io.ReadCloser // RC of zipped file in archive
		fname   string        // name of file being extracted
		mp3     *os.File      // extracted file
		mp3name string        // name of final file
		mp3st   os.FileInfo   // stat of final file

		err error
	)

	fname = filepath.Base(zf.Name)

	if !utf8.ValidString(fname) {
		if fname, err = iconv.ConvertString(fname, "cp437", "utf-8"); err != nil {
			panic(err)
		}
	}

	mp3name = filepath.Join(*destdir, fname)

	if mp3st, err = os.Stat(mp3name); err == nil && mp3st.Size() == int64(zf.UncompressedSize64) {
		log.Println(" ", fname, "skipped")
		return
	}

	log.Println(" ", fname, humanize.Bytes(zf.UncompressedSize64))

	if mp3, err = os.Create(filepath.Join(*destdir, fname)); err != nil {
		panic(err)
	} else if zfbody, err = zf.Open(); err != nil {
		panic(err)
	}
	defer zfbody.Close()

	if io.Copy(mp3, zfbody); err != nil {
		panic(err)
	}
}

func main() {
	flag.Parse()

	for _, dir := range []string{*destdir, *tmpdir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			panic(err)
		}
	}

	for zf := range zipfiles() {
		extract(zf)
	}
}
