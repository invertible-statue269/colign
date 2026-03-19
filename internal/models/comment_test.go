package models

import "testing"

func TestCommentModel(t *testing.T) {
	c := &Comment{
		DocumentID: 1,
		UserID:     1,
		Content:    "Needs clarification",
		RangeFrom:  10,
		RangeTo:    25,
		Resolved:   false,
	}

	if c.Resolved {
		t.Error("new comment should not be resolved")
	}
}

func TestCommentReplyModel(t *testing.T) {
	r := &CommentReply{
		CommentID: 1,
		UserID:    2,
		Content:   "Fixed it",
	}

	if r.CommentID != 1 {
		t.Errorf("expected comment_id 1, got %d", r.CommentID)
	}
}
