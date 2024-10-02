package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

// OPML構造
type OPML struct {
	XMLName xml.Name `xml:"opml"`
	Version string   `xml:"version,attr"`
	Head    Head     `xml:"head"`
	Body    Body     `xml:"body"`
}

type Head struct {
	Title string `xml:"title"`
}

type Body struct {
	Outline []Outline `xml:"outline"`
}

type Outline struct {
	Text     string    `xml:"text,attr"`
	Children []Outline `xml:"outline,omitempty"`
}

// PlantUMLからOPMLに変換
func plantUMLToOPML(plantUML string) OPML {
	lines := strings.Split(plantUML, "\n")
	rootOutline := Outline{}
	stack := []*Outline{&rootOutline}

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "@start") || strings.HasPrefix(line, "@end") {
			continue
		}

		level := strings.Count(line, "*")
		title := strings.TrimSpace(line[level:])
		outline := Outline{Text: title}

		if level > len(stack)-1 {
			parent := stack[len(stack)-1]
			parent.Children = append(parent.Children, outline)
			stack = append(stack, &parent.Children[len(parent.Children)-1])
		} else {
			stack = stack[:level]
			parent := stack[len(stack)-1]
			parent.Children = append(parent.Children, outline)
			stack = append(stack, &parent.Children[len(parent.Children)-1])
		}
	}

	return OPML{
		Version: "2.0",
		Head:    Head{Title: "Converted Mindmap"},
		Body:    Body{Outline: rootOutline.Children},
	}
}

// OPMLからPlantUMLに変換
func opmlToPlantUML(opml OPML) string {
	var builder strings.Builder
	builder.WriteString("@startmindmap\n")

	var traverse func(outline Outline, level int)
	traverse = func(outline Outline, level int) {
		builder.WriteString(strings.Repeat("*", level) + " " + outline.Text + "\n")
		for _, child := range outline.Children {
			traverse(child, level+1)
		}
	}

	for _, outline := range opml.Body.Outline {
		traverse(outline, 1)
	}

	builder.WriteString("@endmindmap")
	return builder.String()
}

// OPMLファイルの読み込み
func loadOPML(filename string) (OPML, error) {
	file, err := os.Open(filename)
	if err != nil {
		return OPML{}, err
	}
	defer file.Close()

	var opml OPML
	if err := xml.NewDecoder(file).Decode(&opml); err != nil {
		return OPML{}, err
	}

	return opml, nil
}

// OPMLファイルの保存
func saveOPML(filename string, opml OPML) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := xml.NewEncoder(file)
	encoder.Indent("", "  ")
	if err := encoder.Encode(opml); err != nil {
		return err
	}

	return nil
}

// ファイルの内容判定
func detectFileType(content string) string {
	if strings.HasPrefix(strings.TrimSpace(content), "@startmindmap") {
		return "plantuml"
	}
	if strings.Contains(content, "<opml") {
		return "opml"
	}
	return "unknown"
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: converter <filename>")
		return
	}

	inputFile := os.Args[1]
	contentBytes, err := ioutil.ReadFile(inputFile)
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}

	content := string(contentBytes)
	fileType := detectFileType(content)

	if fileType == "plantuml" {
		opml := plantUMLToOPML(content)
		outputFile := strings.TrimSuffix(inputFile, ".plantuml") + ".opml"
		err := saveOPML(outputFile, opml)
		if err != nil {
			fmt.Println("Error saving OPML file:", err)
		} else {
			fmt.Println("Converted PlantUML to OPML:", outputFile)
		}
	} else if fileType == "opml" {
		opml, err := loadOPML(inputFile)
		if err != nil {
			fmt.Println("Error loading OPML file:", err)
			return
		}
		plantUML := opmlToPlantUML(opml)
		outputFile := strings.TrimSuffix(inputFile, ".opml") + ".plantuml"
		err = ioutil.WriteFile(outputFile, []byte(plantUML), 0644)
		if err != nil {
			fmt.Println("Error saving PlantUML file:", err)
		} else {
			fmt.Println("Converted OPML to PlantUML:", outputFile)
		}
	} else {
		fmt.Println("Unknown file type.")
	}
}
