package character

import (
	"bufio"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"warofages/internal/util"
	"warofages/internal/woa"
)

func CharacterHandler(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		// No ID provided — serve sessions list
		characters(w, r)
		return
	} else {
		characterDetailHandler(w, r)
	}
}

func characters(w http.ResponseWriter, r *http.Request) {
	characters, err := getCharacters()
	if err != nil {
		return
	}
	tmpl, err := template.ParseFiles("static/characters/index.html")
	if err != nil {
		return
	}
	tmpl.Execute(w, characters)
}

func characterDetailHandler(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")

	characterName := name

	tmpl, err := template.ParseFiles("static/characters/character.html")
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}

	characters, _ := getCharacters()

	var selected woa.Character
	for _, a := range characters {
		if a.Name == characterName {
			selected = a
		}
	}

	tmpl.Execute(w, selected)
}

func loadCharacterMarkdown(path string) (woa.Character, error) {
	file, err := os.Open(path)
	if err != nil {
		return woa.Character{}, err
	}
	defer file.Close()

	var c woa.Character
	var mdLines []string
	scanner := bufio.NewScanner(file)
	inMeta := false
	metaStarted := false

	for scanner.Scan() {
		line := scanner.Text()

		if strings.TrimSpace(line) == "---" {
			if !metaStarted {
				// First --- encountered, start reading metadata
				inMeta = true
				metaStarted = true
				continue
			} else if inMeta {
				// Second --- encountered, stop reading metadata
				inMeta = false
				continue
			}
		}

		if inMeta {
			if parts := strings.SplitN(line, ":", 2); len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				val := strings.TrimSpace(parts[1])
				switch key {
				case "Name":
					c.Name = val
				case "Race":
					c.Race = val
				case "Class":
					c.Class = val
				case "Age":
					c.Age = val
				case "Level":
					c.Level = val
				}
			}
		} else if metaStarted {
			// Only collect markdown lines after metadata section has ended
			mdLines = append(mdLines, line)
		}
	}

	md := strings.Join(mdLines, "\n")
	c.Body = util.MdToHTML([]byte(md))
	return c, nil
}

func getCharacters() ([]woa.Character, error) {
	files, err := filepath.Glob("./md/chars/*.md")
	if err != nil {
		return nil, err
	}
	result := make([]woa.Character, len(files))

	for i, file := range files {
		result[i], _ = loadCharacterMarkdown(file)
	}
	return result, nil
}
