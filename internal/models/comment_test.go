package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCommentModel(t *testing.T) {
	c := &Comment{
		ChangeID:     1,
		DocumentType: "proposal",
		QuotedText:   "selected text",
		Body:         "Needs clarification",
		UserID:       1,
		Resolved:     false,
	}

	assert.False(t, c.Resolved, "new comment should not be resolved")
}

func TestCommentReplyModel(t *testing.T) {
	r := &CommentReply{
		CommentID: 1,
		UserID:    2,
		Body:      "Fixed it",
	}

	assert.Equal(t, int64(1), r.CommentID)
}
