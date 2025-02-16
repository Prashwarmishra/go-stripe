package main

import (
	"embed"
	"fmt"
	"net/http"
	"strings"
	"text/template"
)

type templateData struct {
	StringMap       map[string]string
	IntMap          map[string]int
	FloatMap        map[string]float64
	Data            map[string]any
	Flash           string
	Warning         string
	Error           string
	CSRFToken       string
	IsAuthenticated int
	API             string
	CSSVersion      string
}

var functions = template.FuncMap{}

//go:embed templates
var templatesFS embed.FS

func (app *application) getDefaultTemplateData(td *templateData, r *http.Request) *templateData {
	return td
}

func (app *application) renderTemplate(w http.ResponseWriter, r *http.Request, page string, td *templateData, partials ...string) error {
	var t *template.Template
	var err error

	templateToRender := fmt.Sprintf("templates/%s.page.gohtml", page)

	templateCache, inTemplateCache := app.templateCache[templateToRender]

	if inTemplateCache && app.config.env == "production" {
		t = templateCache
	} else {
		t, err = app.parseTemplate(page, templateToRender, partials)

		if err != nil {
			app.errorLog.Println("error setting template", err)
			return err
		}
	}

	if td == nil {
		td = &templateData{}
	}

	td = app.getDefaultTemplateData(td, r)

	err = t.Execute(w, td)

	if err != nil {
		app.errorLog.Println("error rendering template", err)
		return err
	}
	return nil
}

func (app *application) parseTemplate(page, templateToRender string, partials []string) (*template.Template, error) {
	var t *template.Template
	var err error

	if len(partials) > 0 {
		for index, partial := range partials {
			partials[index] = fmt.Sprintf("templates/%s.partial.gohtml", partial)
		}

		t, err = template.New(fmt.Sprintf("%s.page.gohtml", page)).Funcs(functions).ParseFS(templatesFS, "templates/base.layout.gohtml", strings.Join(partials, ","), templateToRender)
	} else {
		t, err = template.New(fmt.Sprintf("%s.page.gohtml", page)).Funcs(functions).ParseFS(templatesFS, "templates/base.layout.gohtml", templateToRender)

	}

	if err != nil {
		app.errorLog.Println("error parsing template:", err)
		return nil, err
	}

	app.templateCache[templateToRender] = t
	return t, nil
}
