package markdown

import (
	"fmt"
	"html"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"

	lineTooLongRender "github.com/MichaelMure/go-term-markdown"
	"github.com/charmbracelet/glamour"
	"github.com/stroborobo/aimg"
)

type Model struct {
	Content  string
	Width    int
	RootPath string
}

func (m Model) Render() string {
	content, imageMap := imagePlaceholders(m.Content)

	glamourRender, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithEmoji(),
		glamour.WithPreservedNewLines(),
	)

	if err != nil {
		log.Fatal(err)
	}

	markdown, err := glamourRender.Render(content)

	if err != nil {
		log.Fatal(err)
	}

	markdown = renderLinesTooLong(markdown, m.Width)

	// Remplacer les identifiants d'images par les images elles-mêmes
	for placeholder, img := range imageMap {
		if isURL(img) {
			ansiImage, ok := m.renderImageFromUrl(img)
			if ok {
				markdown = strings.Replace(markdown, placeholder, ansiImage, -1)
			}
		} else {
			ansiImage, ok := m.renderImageLocal(img)
			if ok {
				markdown = strings.Replace(markdown, placeholder, ansiImage, -1)
			}
		}
	}

	return markdown
}

// imagePlaceholders trouve toutes les images au format Markdown dans content,
// les remplace par des identifiants uniques, et retourne le nouveau contenu
// ainsi qu'une map associant les identifiants aux URLs des images.
func imagePlaceholders(content string) (string, map[string]string) {
	// Expression régulière pour matcher les images Markdown
	re := regexp.MustCompile(`!\[.*?\]\((.*?)\)`)

	// Map pour stocker les correspondances entre identifiants et URLs d'images
	imageMap := make(map[string]string)

	// Remplacer chaque image trouvée par un identifiant unique
	newContent := re.ReplaceAllStringFunc(content, func(match string) string {
		// Extraire l'URL de l'image
		url := re.FindStringSubmatch(match)[1]
		// Générer un identifiant unique pour l'image
		placeholder := fmt.Sprintf("{{image%d}}", len(imageMap)+1)
		// Associer l'identifiant à l'URL dans la map
		imageMap[placeholder] = url
		// Retourner l'identifiant pour le remplacer dans le contenu
		return placeholder
	})

	return newContent, imageMap
}

func (m Model) renderImageFromUrl(url string) (string, bool) {
	resp, err := http.Get(url)

	if err != nil {
		return "", false
	}

	defer resp.Body.Close()

	contentType := resp.Header.Get("Content-Type")
	if strings.HasPrefix(contentType, "image/") {
		img := aimg.NewImage(m.Width - (m.Width / 2))
		img.ParseReader(resp.Body)

		ansiImage := addPaddingToLeft(img.String())

		return ansiImage, true
	}

	return "", false
}

func (m Model) renderImageLocal(path string) (string, bool) {
	if m.RootPath != "" {
		path = m.RootPath + "/" + path
	}

	file, err := os.Open(path)

	if err != nil {
		return "", false
	}

	defer file.Close()

	img := aimg.NewImage(m.Width - (m.Width / 2))
	img.ParseReader(file)

	ansiImage := addPaddingToLeft(img.String())

	return ansiImage, true
}

func isURL(str string) bool {
	u, err := url.Parse(str)
	return err == nil && u.Scheme != "" && u.Host != ""
}

func addPaddingToLeft(input string) string {
	// Séparer l'input en lignes
	lines := strings.Split(input, "\n")

	// Préparer un slice pour stocker les nouvelles lignes
	paddedLines := make([]string, len(lines))

	// Ajouter deux espaces à chaque ligne
	for i, line := range lines {
		if i == 0 {
			paddedLines[i] = strings.TrimSpace(line)
			continue
		}
		paddedLines[i] = "  " + strings.TrimSpace(line)
	}

	// Rejoindre les lignes modifiées en une seule chaîne
	return strings.Join(paddedLines, "\n")
}

func renderLinesTooLong(content string, width int) string {
	lines := strings.Split(content, "\n")
	outputLines := []string{}
	for _, line := range lines {
		realString := realString(line)
		if len(realString) >= width {
			line = addPaddingToLeft(string(lineTooLongRender.Render(line, width-5, 0)))
		}
		outputLines = append(outputLines, line)
	}

	outputText := strings.Join(outputLines, "\n")

	return outputText
}

func realString(input string) string {
	result := removeANSIEscapeSequences(input)

	// Convert the string to HTML entities
	result = html.EscapeString(result)

	// Unescape the HTML entities to clean the string
	result = html.UnescapeString(result)

	return result
}

func removeANSIEscapeSequences(s string) string {
	re := regexp.MustCompile("\x1b\\[[0-9;]*m")
	return strings.TrimSpace(re.ReplaceAllString(s, ""))
}
