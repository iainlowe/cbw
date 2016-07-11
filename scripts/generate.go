package main

import (
	"bufio"
	"html/template"
	"io/ioutil"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
)

func extractNum(s string) string {
	parts := strings.Split(s, " ")
	num := strings.Replace(parts[len(parts)-1], ".mp3", "", -1)

	if num == "75c" { // Handle weird case
		num = "75"
	}

	return num
}

// Implement length-based sort with ByLen type.
type ByNum []string

func (a ByNum) Len() int { return len(a) }
func (a ByNum) Less(i, j int) bool {
	si, sj := extractNum(a[i]), extractNum(a[j])

	if ni, err := strconv.ParseInt(si, 10, 32); err == nil {
		if nj, err := strconv.ParseInt(sj, 10, 32); err == nil {
			return ni < nj
		}
	}

	sorted := []string{si, sj}

	sort.Strings(sorted)

	return sorted[0] == si
}
func (a ByNum) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

func getSongs() (songs []Song) {
	finfos, _ := ioutil.ReadDir("mp3")
	names := []string{}

	for _, info := range finfos {
		names = append(names, info.Name())
	}

	sort.Sort(ByNum(names))

	for _, n := range names {
		songs = append(songs, NewSong(n))
	}

	return
}

type Song struct {
	Path     string
	Name     string
	Number   string
	Category string
}

func (s Song) String() string {
	return s.Path
}

func NewSong(song string) Song {
	s := Song{
		Path: "mp3/" + song,
	}

	s.Number = extractNum(song)

	if s.Number == "75" { // Handle weird case
		s.Number = "75C"
	}

	parts := strings.Split(song, " ")
	s.Category = strings.Join(parts[1:len(parts)-1], " ")

	if nm, ok := song_index[s.Number]; ok {
		s.Name = nm
	} else {
		s.Name = "N/A"
	}

	return s
}

var song_index = map[string]string{}

func loadSongIndex() {
	file, err := os.Open("data/songindex_b")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		p := strings.Split(line, " ")
		num := p[0]
		song_index[num] = strings.Join(p[1:], " ")
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}

func main() {
	loadSongIndex()

	tpl := template.Must(template.ParseFiles("data/tpl.html"))

	if f, err := os.Create("index.html"); err != nil {
		log.Fatal(err)
	} else if err := tpl.Execute(f, getSongs()); err != nil {
		log.Fatal(err)
	}
}
