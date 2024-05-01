package main

import (
	"fmt"
	"os"

	"github.com/go-scripts/markdown"
)

func main() {
	file, err := os.ReadFile("./markdown.md")
	if err != nil {
		panic(err)
	}

	content := string(file)
	width := 80
	Markdown := markdown.Model{Content: content, Width: width}
	result := Markdown.Render()

	fmt.Println(result)
}
