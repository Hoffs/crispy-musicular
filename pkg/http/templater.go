package http

import (
	"html/template"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog/log"
)

type templater struct {
	cacheTemplates bool
	cache          map[string]*template.Template
	directory      string
	layouts        *template.Template // use loadLayouts instead in code that doesn't actually load layouts
}

// Could return interface instead, but since this is not exported I am unsure if there is a benefit.
func NewTemplater(directory string, cacheTemplates bool) *templater {
	t := &templater{
		cacheTemplates: cacheTemplates,
		cache:          make(map[string]*template.Template),
		directory:      directory,
	}

	if t.cacheTemplates {
		t.preloadTemplates()
	}

	return t
}

// Attempts to render a template
// Template has to exist in specified directory and must have
// a defined block called "entrypoint" which will be used to execute template.
func (t *templater) renderTemplate(w http.ResponseWriter, template string, data interface{}) {
	log.Debug().Msgf("templater: trying to render %s", template)

	tmpl, err := t.loadTemplate(template)
	if err != nil {
		log.Error().Err(err).Msgf("templater: failed to load template %s", template)
		http.Error(w, http.StatusText(500), 500)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	err = tmpl.ExecuteTemplate(w, "entrypoint", data)
	if err != nil {
		log.Error().Err(err).Msgf("templater: failed to render template %s", template)
		http.Error(w, http.StatusText(500), 500)
	}
}

// This will populate cache, but also returns template as well
func (t *templater) loadTemplate(name string) (tmpl *template.Template, err error) {
	if tmpl, ok := t.cache[name]; ok {
		return tmpl, nil
	}

	log.Debug().Msgf("templater: loading template %s", name)
	layouts, err := t.loadLayouts()
	if err != nil {
		return nil, err
	}

	layouts, err = layouts.Clone()
	if err != nil {
		return nil, err
	}

	tmpl, err = layouts.ParseFiles(filepath.Join("templates", name))

	if t.cacheTemplates {
		t.cache[name] = tmpl
	}

	return
}

func (t *templater) loadLayouts() (tmpl *template.Template, err error) {
	if t.layouts != nil {
		return t.layouts, nil
	}

	log.Debug().Msg("templater: loading layouts")
	tmpl, err = template.ParseGlob(filepath.Join(t.directory, "*.layout.tmpl"))
	if err != nil {
		return nil, err
	}

	if t.cacheTemplates {
		t.layouts = tmpl
	}

	return
}

func (t *templater) preloadTemplates() (err error) {
	layouts, err := t.loadLayouts()
	if err != nil {
		return
	}

	log.Debug().Msg("templater: loading all templates")
	var layoutFiles []string
	for _, template := range layouts.Templates() {
		name := template.Name()
		if strings.HasSuffix(name, "layout.tmpl") {
			layoutFiles = append(layoutFiles, name)
		}
	}

	// Iterate over non-"layout" files and cache them
	files, err := filepath.Glob(filepath.Join(t.directory, "*.tmpl"))
FILE_LOAD_LOOP:
	for _, file := range files {
		filename := filepath.Base(file)
		if filename == "." || filename == string(filepath.Separator) {
			continue
		}

		// Could use a map, but this should be small enough that it should not matter
		isLayout := false
		for _, layoutName := range layoutFiles {
			isLayout = layoutName == filename
			if isLayout {
				// Break out and go to next file
				continue FILE_LOAD_LOOP
			}
		}

		layouts, err := layouts.Clone()
		if err != nil {
			return err
		}

		fileTemplate, err := layouts.ParseFiles(file)
		if err != nil {
			return err
		}

		t.cache[filename] = fileTemplate
		log.Debug().Msgf("templater: loaded %s", filename)
	}

	return
}
