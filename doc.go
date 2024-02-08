package gocfg

type DocField struct {
	Key          string
	Description  string
	DefaultValue string
	OmitEmpty    bool
}

type DocTree struct {
	Title  string
	Fields []*DocField
	Groups []*DocTree
}

func NewDoc() *DocTree {
	return &DocTree{
		Fields: []*DocField{},
		Groups: []*DocTree{},
	}
}

func (d *DocTree) AddGroup(name string) *DocTree {
	g := &DocTree{
		Title:  name,
		Fields: []*DocField{},
		Groups: []*DocTree{},
	}

	d.Groups = append(d.Groups, g)

	return g
}

func (d *DocTree) AddField(field *DocField) *DocField {
	d.Fields = append(d.Fields, field)

	return field
}
