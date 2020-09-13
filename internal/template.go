package internal

import (
	"bytes"
	"html/template"

	"github.com/Masterminds/sprig"
	"gopkg.in/alecthomas/kingpin.v2"
	"gopkg.in/yaml.v2"
)

func ExecuteYAMLTemplate(name string, i interface{}, templateVars map[string]interface{}) (*bytes.Buffer, error) {
	y, err := yaml.Marshal(i)
	if err != nil {
		kingpin.Errorf("could not marshal template, %v", err)
		return nil, err
	}

	t := template.Must(template.New(name).
		Delims("$((", "))").
		Funcs(sprig.FuncMap()).
		Parse(string(y)))

	bytesCache := new(bytes.Buffer)
	err = t.Execute(bytesCache, templateVars)
	if err != nil {
		kingpin.Errorf("could not template, %v", err)
		return nil, err
	}

	return bytesCache, nil
}
