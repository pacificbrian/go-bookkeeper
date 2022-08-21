/*
 * SPDX-FileCopyrightText: 2022 Brian Welty
 *
 * SPDX-License-Identifier: MPL-2.0
 */

package route

import (
	"errors"
	"html/template"
	"io"
	"github.com/labstack/echo/v4"
)

type Template struct {
	// map[key] is template to be called by controller/action
	templates map[string]*template.Template
}

func NewTemplate() *Template {
	return &Template{
		templates: make(map[string]*template.Template),
	}
}

// Used by echo framework. Don't use this function directly.
func (t *Template) Render(w io.Writer, template_key string, data interface{}, c echo.Context) error {
	if tmpl, exist := t.templates[template_key]; exist {
		 // This wll execute the map[key] template file
		return tmpl.Execute(w, data)
	} else {
		return errors.New("Template map[" + template_key + "] is missing.")
	}
}

func (tmpl *Template) Add(template_key string, view_file_name string) {
	tmpl.templates[template_key] = template.Must(template.ParseFiles(view_file_name))
}
