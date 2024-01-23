package generator

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"os"
	"strings"
)

type Generator interface {
	Generate(ctx context.Context, template []byte) ([]byte, error)
}

type EnvironmentGenerator struct {
	envs map[string]string
}

func NewEnvironmentGenerator() *EnvironmentGenerator {
	res := make(map[string]string)
	for _, v := range os.Environ() {
		pair := strings.SplitN(v, "=", 2)
		res[pair[0]] = pair[1]
	}

	return &EnvironmentGenerator{
		envs: res,
	}
}

func (e *EnvironmentGenerator) Generate(_ context.Context, t []byte) ([]byte, error) {
	tmpl, err := template.New("template").Parse(string(t))
	if err != nil {
		return nil, fmt.Errorf("EnvironmentFileGenerator::Generate error while parsing template %w", err)
	}
	res := new(bytes.Buffer)
	if err := tmpl.Execute(res, e.envs); err != nil {
		return nil, fmt.Errorf("EnvironmentFileGenerator::Generate error while executing template %w", err)
	}
	return res.Bytes(), nil
}
