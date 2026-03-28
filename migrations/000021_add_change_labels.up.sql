CREATE TABLE change_label_assignments (
    change_id BIGINT NOT NULL REFERENCES changes(id) ON DELETE CASCADE,
    label_id BIGINT NOT NULL REFERENCES project_labels(id) ON DELETE CASCADE,
    PRIMARY KEY (change_id, label_id)
);

CREATE INDEX idx_change_label_assignments_label_id ON change_label_assignments(label_id);
