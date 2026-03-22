package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDocumentTypes(t *testing.T) {
	types := []DocumentType{DocProposal, DocDesign, DocTasks}
	expected := []string{"proposal", "design", "tasks"}

	for i, dt := range types {
		assert.Equal(t, expected[i], string(dt))
	}
}

func TestDocumentModel(t *testing.T) {
	doc := &Document{
		ChangeID: 1,
		Type:     DocProposal,
		Content:  "## Why\n\nTest content",
	}

	assert.Equal(t, DocProposal, doc.Type)
}
