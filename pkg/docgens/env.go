package docgens

import (
	"fmt"
	"github.com/Jagerente/gocfg"
	"io"
	"strings"
)

type EnvDocGenerator struct {
	writer io.Writer
}

func NewEnvDocGenerator(writer io.Writer) *EnvDocGenerator {
	return &EnvDocGenerator{
		writer: writer,
	}
}

func (g *EnvDocGenerator) GenerateDoc(doc *gocfg.DocTree) error {
	if err := g.write("# Auto-generated config\n"); err != nil {
		return err
	}

	for _, field := range doc.Fields {
		if err := g.writeField(field); err != nil {
			return err
		}
	}

	for _, group := range doc.Groups {
		if err := g.writeGroup(group); err != nil {
			return err
		}
	}

	return nil
}

func (g *EnvDocGenerator) buildHeader(text string) string {
	return fmt.Sprintf("#############################\n# %s\n#############################\n", text)
}

func (g *EnvDocGenerator) write(text string) error {
	if _, err := fmt.Fprint(g.writer, text); err != nil {
		return fmt.Errorf("failed to write: %w", err)
	}

	return nil
}

func (g *EnvDocGenerator) writeGroup(group *gocfg.DocTree) error {
	if err := g.writeBreakLine(); err != nil {
		return err
	}

	if group.Title != "" {
		if err := g.write(g.buildHeader(group.Title)); err != nil {
			return err
		}
	}

	for _, field := range group.Fields {
		if err := g.writeField(field); err != nil {
			return err
		}
	}

	for _, innerGroup := range group.Groups {
		if err := g.writeGroup(innerGroup); err != nil {
			return err
		}
	}

	return nil
}

func (g *EnvDocGenerator) writeField(field *gocfg.DocField) error {
	if err := g.writeBreakLine(); err != nil {
		return err
	}

	if field.OmitEmpty {
		if err := g.write("# Allowed to be empty\n"); err != nil {
			return err
		}
	}

	if field.Description != "" {
		if err := g.write("# Description:\n"); err != nil {
			return err
		}

		lines := strings.Split(field.Description, "\n")
		for _, line := range lines {
			if err := g.write(fmt.Sprintf("#  %s\n", line)); err != nil {
				return err
			}
		}
	}

	if field.Key != "" {
		if err := g.write(fmt.Sprintf("%s=%s\n", field.Key, field.DefaultValue)); err != nil {
			return err
		}
	}

	return nil
}

func (g *EnvDocGenerator) writeBreakLine() error {
	return g.write("\n")
}
