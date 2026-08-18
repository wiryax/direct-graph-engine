package file

import (
	"html/template"

	dge "github.com/wiryax/dage/core"
)

type HTMLTemplateConnector struct {
	templateFile string
	htmlTemplate *template.Template
}

func NewHTMLTemplateConnector(template string) *HTMLTemplateConnector {
	return &HTMLTemplateConnector{
		templateFile: template,
	}
}

func (h *HTMLTemplateConnector) Validate(_ *dge.GraphContext) error {
	template, err := template.ParseFiles(h.templateFile)
	if err != nil {
		return err
	}

	h.htmlTemplate = template
	return nil
}

func (h *HTMLTemplateConnector) Acquire(_ any) any {
	return h.htmlTemplate
}

func (h *HTMLTemplateConnector) Release() error {
	return nil
}
