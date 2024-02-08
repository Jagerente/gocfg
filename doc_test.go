package gocfg

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_NewDoc(t *testing.T) {
	doc := NewDoc()

	assert.NotNil(t, doc)
	assert.NotNil(t, doc.Fields)
	assert.NotNil(t, doc.Groups)
	assert.Empty(t, doc.Fields)
	assert.Empty(t, doc.Groups)
}

func Test_AddGroup(t *testing.T) {
	doc := NewDoc()

	group1 := doc.AddGroup("Group 1")
	assert.NotNil(t, group1)
	assert.Equal(t, "Group 1", group1.Title)
	assert.Empty(t, group1.Fields)
	assert.Empty(t, group1.Groups)

	group2 := doc.AddGroup("Group 2")
	assert.NotNil(t, group2)
	assert.Equal(t, "Group 2", group2.Title)
	assert.Empty(t, group2.Fields)
	assert.Empty(t, group2.Groups)

	assert.Len(t, doc.Groups, 2)
	assert.Equal(t, group1, doc.Groups[0])
	assert.Equal(t, group2, doc.Groups[1])
}

func Test_AddField(t *testing.T) {
	doc := NewDoc()

	field1 := &DocField{Key: "Field 1", Description: "Description 1", DefaultValue: "Default 1", OmitEmpty: true}
	doc.AddField(field1)
	assert.NotNil(t, doc.Fields)
	assert.Len(t, doc.Fields, 1)
	assert.Equal(t, field1, doc.Fields[0])

	field2 := &DocField{Key: "Field 2", Description: "Description 2", DefaultValue: "Default 2", OmitEmpty: false}
	doc.AddField(field2)
	assert.Len(t, doc.Fields, 2)
	assert.Equal(t, field2, doc.Fields[1])
}

func Test_DocTree(t *testing.T) {
	doc := NewDoc()

	group1 := doc.AddGroup("Group 1")
	group1.AddField(&DocField{Key: "Field 1", Description: "Description 1", DefaultValue: "Default 1", OmitEmpty: true})

	group2 := doc.AddGroup("Group 2")
	group2.AddField(&DocField{Key: "Field 2", Description: "Description 2", DefaultValue: "Default 2", OmitEmpty: false})

	assert.Len(t, doc.Groups, 2)
	assert.Equal(t, "Group 1", doc.Groups[0].Title)
	assert.Len(t, doc.Groups[0].Fields, 1)
	assert.Equal(t, "Field 1", doc.Groups[0].Fields[0].Key)
	assert.Equal(t, "Description 1", doc.Groups[0].Fields[0].Description)
	assert.Equal(t, "Default 1", doc.Groups[0].Fields[0].DefaultValue)
	assert.True(t, doc.Groups[0].Fields[0].OmitEmpty)

	assert.Equal(t, "Group 2", doc.Groups[1].Title)
	assert.Len(t, doc.Groups[1].Fields, 1)
	assert.Equal(t, "Field 2", doc.Groups[1].Fields[0].Key)
	assert.Equal(t, "Description 2", doc.Groups[1].Fields[0].Description)
	assert.Equal(t, "Default 2", doc.Groups[1].Fields[0].DefaultValue)
	assert.False(t, doc.Groups[1].Fields[0].OmitEmpty)
}
