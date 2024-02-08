package docgens

import (
	"bytes"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Jagerente/gocfg"
)

func TestNewEnvDocGenerator(t *testing.T) {
	envDocGen := NewEnvDocGenerator(bytes.NewBuffer(nil))
	assert.NotNil(t, envDocGen)
}

func TestEnvDocGenerator_GenerateDoc(t *testing.T) {
	doc := &gocfg.DocTree{
		Title: "TestDoc",
		Fields: []*gocfg.DocField{
			{Key: "FIELD_1_ENV", Description: "Description for Field1", DefaultValue: "qwe"},
			{Key: "FIELD_2_ENV", Description: "Description for Field2"},
			{Key: "FIELD_3_ENV", Description: "Multi\nLine\nDescription for Field3", OmitEmpty: true},
			{Key: "FIELD_4_ENV", DefaultValue: "123"},
		},
		Groups: []*gocfg.DocTree{
			{Title: "Group1", Fields: []*gocfg.DocField{{Key: "FIELD_2_ENV"}}},
		},
	}

	var buf = new(bytes.Buffer)
	envDocGen := NewEnvDocGenerator(buf)

	err := envDocGen.GenerateDoc(doc)
	assert.NoError(t, err)

	expectedOutput := `# Auto-generated config

# Description:
#  Description for Field1
FIELD_1_ENV=qwe

# Description:
#  Description for Field2
FIELD_2_ENV=

# Allowed to be empty
# Description:
#  Multi
#  Line
#  Description for Field3
FIELD_3_ENV=

FIELD_4_ENV=123

#############################
# Group1
#############################

FIELD_2_ENV=
`

	assert.Equal(t, expectedOutput, buf.String())
}

func TestEnvDocGenerator_GenerateDoc_ErrorOnWrite(t *testing.T) {
	failingWriter := &mockFailingWriter{
		failAfter: 0,
	}

	envDocGen := &EnvDocGenerator{writer: failingWriter}

	err := envDocGen.GenerateDoc(new(gocfg.DocTree))
	assert.Error(t, err)
}

func TestEnvDocGenerator_GenerateDoc_ErrorOnWriteField(t *testing.T) {
	doc := &gocfg.DocTree{
		Fields: []*gocfg.DocField{
			{Key: "FIELD_1_ENV", Description: "Description for Field1", DefaultValue: "qwe"},
		},
	}

	failingWriter := &mockFailingWriter{
		failAfter: 1,
	}

	envDocGen := &EnvDocGenerator{writer: failingWriter}

	err := envDocGen.GenerateDoc(doc)
	assert.Error(t, err)
}

func TestEnvDocGenerator_GenerateDoc_ErrorOnWriteGroup(t *testing.T) {
	doc := &gocfg.DocTree{
		Groups: []*gocfg.DocTree{
			{Title: "Group1", Fields: []*gocfg.DocField{{Key: "FIELD_2_ENV"}}},
		},
	}

	failingWriter := &mockFailingWriter{
		failAfter: 1,
	}

	envDocGen := &EnvDocGenerator{writer: failingWriter}

	err := envDocGen.GenerateDoc(doc)
	assert.Error(t, err)
}

func TestEnvDocGenerator_writeGroup_ErrorOnBreakLine(t *testing.T) {
	failingWriter := &mockFailingWriter{
		failAfter: 0,
	}

	envDocGen := &EnvDocGenerator{writer: failingWriter}

	err := envDocGen.writeGroup(new(gocfg.DocTree))
	assert.Error(t, err)
}

func TestEnvDocGenerator_writeGroup_ErrorOnBuildHeader(t *testing.T) {
	doc := &gocfg.DocTree{
		Title: "TestDoc",
	}

	failingWriter := &mockFailingWriter{
		failAfter: 1,
	}

	envDocGen := &EnvDocGenerator{writer: failingWriter}

	err := envDocGen.writeGroup(doc)
	assert.Error(t, err)
}

func TestEnvDocGenerator_writeGroup_ErrorOnWriteField(t *testing.T) {
	doc := &gocfg.DocTree{
		Fields: []*gocfg.DocField{
			{Key: "FIELD_1_ENV", Description: "Description for Field1", DefaultValue: "qwe"},
		},
	}

	failingWriter := &mockFailingWriter{
		failAfter: 1,
	}

	envDocGen := &EnvDocGenerator{writer: failingWriter}

	err := envDocGen.writeGroup(doc)
	assert.Error(t, err)
}

func TestEnvDocGenerator_writeGroup_ErrorOnWriteGroup(t *testing.T) {
	doc := &gocfg.DocTree{
		Groups: []*gocfg.DocTree{
			{Title: "Group1", Fields: []*gocfg.DocField{{Key: "FIELD_2_ENV"}}},
		},
	}

	failingWriter := &mockFailingWriter{
		failAfter: 1,
	}

	envDocGen := &EnvDocGenerator{writer: failingWriter}

	err := envDocGen.writeGroup(doc)
	assert.Error(t, err)
}

func TestEnvDocGenerator_writeGroup_NoError(t *testing.T) {
	doc := &gocfg.DocTree{
		Title:  "TestDoc",
		Fields: []*gocfg.DocField{},
		Groups: []*gocfg.DocTree{
			{Title: "Group1", Fields: []*gocfg.DocField{{Key: "FIELD_2_ENV"}}},
		},
	}

	envDocGen := &EnvDocGenerator{writer: bytes.NewBuffer(nil)}

	err := envDocGen.writeGroup(doc)
	assert.Nil(t, err)
}

func TestEnvDocGenerator_writeField_ErrorOnBreakLine(t *testing.T) {
	failingWriter := &mockFailingWriter{
		failAfter: 0,
	}

	envDocGen := &EnvDocGenerator{writer: failingWriter}

	err := envDocGen.writeField(new(gocfg.DocField))
	assert.Error(t, err)
}

func TestEnvDocGenerator_writeField_ErrorOnWriteOmitEmpty(t *testing.T) {
	field := &gocfg.DocField{
		OmitEmpty: true,
	}

	failingWriter := &mockFailingWriter{
		failAfter: 1,
	}

	envDocGen := &EnvDocGenerator{writer: failingWriter}

	err := envDocGen.writeField(field)
	assert.Error(t, err)
}

func TestEnvDocGenerator_writeField_ErrorOnWriteDescription(t *testing.T) {
	field := &gocfg.DocField{
		Description: "qwe",
	}

	failingWriter := &mockFailingWriter{
		failAfter: 1,
	}

	envDocGen := &EnvDocGenerator{writer: failingWriter}

	err := envDocGen.writeField(field)
	assert.Error(t, err)
}

func TestEnvDocGenerator_writeField_ErrorOnWriteDescriptionLine(t *testing.T) {
	field := &gocfg.DocField{
		Description: "qwe",
	}

	failingWriter := &mockFailingWriter{
		failAfter: 2,
	}

	envDocGen := &EnvDocGenerator{writer: failingWriter}

	err := envDocGen.writeField(field)
	assert.Error(t, err)
}

func TestEnvDocGenerator_writeField_ErrorOnWriteKey(t *testing.T) {
	field := &gocfg.DocField{
		Key: "qwe",
	}

	failingWriter := &mockFailingWriter{
		failAfter: 1,
	}

	envDocGen := &EnvDocGenerator{writer: failingWriter}

	err := envDocGen.writeField(field)
	assert.Error(t, err)
}

func TestEnvDocGenerator_buildHeader(t *testing.T) {
	envDocGen := &EnvDocGenerator{}

	header := envDocGen.buildHeader("Header")
	expectedOutput := "#############################\n# Header\n#############################\n"
	assert.Equal(t, expectedOutput, header)
}

type mockFailingWriter struct {
	failAfter int
}

func (m *mockFailingWriter) Write(_ []byte) (n int, err error) {
	if m.failAfter != 0 {
		m.failAfter--
		return 0, nil
	}

	return 0, errors.New("mock write error")
}
