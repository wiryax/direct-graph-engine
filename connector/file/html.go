package file

import (
	"html/template"

	dge "github.com/wiryax/dage/core"
)

type HTMLTemplateConnector struct {
	templateFile string
	htmlTemplate *template.Template
}

func (h *HTMLTemplateConnector) Validate(_ *dge.GraphContext) error {
	return nil
}

func (h *HTMLTemplateConnector) Acquire(_ any) any {
	return h.htmlTemplate
}

func (h *HTMLTemplateConnector) Realest() {
}
