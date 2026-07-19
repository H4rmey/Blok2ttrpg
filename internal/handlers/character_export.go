package handlers

import (
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/harmey/blok2ttrpg/ability-builder/internal/export"
	"github.com/harmey/blok2ttrpg/ability-builder/internal/models"
)

// ExportCharacterYAMLHandler returns a YAML file download for an entire
// character including all of their abilities.
func (app *App) ExportCharacterYAMLHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := extractPathParam(r, "characters", 1)
	if id == "" {
		http.NotFound(w, r)
		return
	}

	c, err := app.Store.GetCharacter(id)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	yamlOutput := export.CharacterToYAML(c)

	filename := "character"
	if c.Name != "" {
		filename = sanitizeFilename(c.Name)
	}

	w.Header().Set("Content-Type", "text/yaml; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="`+filename+`.yaml"`)
	w.Write([]byte(yamlOutput))
}

// ImportCharacterYAMLHandler accepts a YAML file upload and creates a new
// character (or merges into an existing one when ?merge={id} is supplied).
func (app *App) ImportCharacterYAMLHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Cap the upload size to a reasonable 1 MiB.
	if err := r.ParseMultipartForm(1 << 20); err != nil {
		http.Error(w, "Failed to parse upload (max 1 MiB)", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("yaml_file")
	if err != nil {
		http.Error(w, "Missing 'yaml_file' upload", http.StatusBadRequest)
		return
	}
	defer file.Close()

	buf := make([]byte, header.Size)
	if _, err := file.Read(buf); err != nil {
		http.Error(w, "Failed to read upload", http.StatusBadRequest)
		return
	}

	character, err := export.ParseCharacterYAML(buf)
	if err != nil {
		http.Error(w, "Failed to parse YAML: "+err.Error(), http.StatusBadRequest)
		return
	}

	// If a merge target was specified, append the imported abilities to it
	// rather than creating a fresh character.
	mergeID := r.URL.Query().Get("merge")
	if mergeID != "" {
		existing, err := app.Store.GetCharacter(mergeID)
		if err != nil {
			http.Error(w, "Target character not found", http.StatusNotFound)
			return
		}
		existing.Abilities = append(existing.Abilities, character.Abilities...)
		if err := app.Store.SaveCharacter(*existing); err != nil {
			http.Error(w, "Failed to save character", http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/characters/"+mergeID+"/abilities", http.StatusSeeOther)
		return
	}

	if err := app.Store.SaveCharacter(*character); err != nil {
		http.Error(w, "Failed to save character", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/characters/"+character.ID, http.StatusSeeOther)
}

// sanitizeFilename strips characters that are unsafe in filenames.
func sanitizeFilename(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return "character"
	}
	ext := filepath.Ext(name)
	base := strings.TrimSuffix(name, ext)
	repl := strings.NewReplacer("/", "-", "\\", "-", ":", "-", "*", "-", "?", "-", "\"", "-", "<", "-", ">", "-", "|", "-")
	return repl.Replace(base) + ext
}

// PdfCharacterHandler renders the character as a print-friendly HTML page
// styled like a D&D/Pathfinder character sheet, with a print button.
func (app *App) PdfCharacterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := extractPathParam(r, "characters", 1)
	if id == "" {
		http.NotFound(w, r)
		return
	}

	c, err := app.Store.GetCharacter(id)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	data := struct {
		Character *models.Character
		ProfDice  map[models.Proficiency]string
	}{
		Character: c,
		ProfDice:  models.ProficiencyDice,
	}

	pdfPath := filepath.Join(app.TemplateDir, "character_pdf.html")
	tmpl, err := template.New("").Funcs(app.funcMap).ParseFiles(pdfPath)
	if err != nil {
		log.Printf("failed to parse character pdf template: %v", err)
		http.Error(w, "Failed to load character sheet template", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteTemplate(w, "character_pdf.html", data); err != nil {
		log.Printf("failed to render character pdf: %v", err)
		http.Error(w, "Failed to render character sheet", http.StatusInternalServerError)
	}
}

var characterPdfTemplate = template.Must(template.New("character_pdf_inline").Parse(`<!DOCTYPE html>
<html lang="en">
<head><meta charset="UTF-8"><title>Character Sheet</title></head>
<body><p>Legacy inline template</p></body>
</html>`))

