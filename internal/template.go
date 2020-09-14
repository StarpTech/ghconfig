package internal

import (
	"bytes"
	"html/template"

	"github.com/Masterminds/sprig"
	"github.com/apex/log"
	"gopkg.in/yaml.v2"
)

func ExecuteYAMLTemplate(name string, i interface{}, templateVars map[string]interface{}) (*bytes.Buffer, error) {
	y, err := yaml.Marshal(i)
	if err != nil {
		log.WithError(err).Error("could not marshal template")
		return nil, err
	}

	t := template.Must(template.New(name).
		Delims("$((", "))").
		Funcs(sprig.FuncMap()).
		Parse(string(y)))

	bytesCache := new(bytes.Buffer)
	err = t.Execute(bytesCache, templateVars)
	if err != nil {
		log.WithError(err).Error("could not execute template")
		return nil, err
	}

	return bytesCache, nil
}
