CREATE TABLE IF NOT EXISTS reactions (
    target_type TEXT NOT NULL CHECK (target_type IN ('topic', 'comment')),
    target_id UUID NOT NULL,
    visitor_id UUID NOT NULL,
    reaction TEXT NOT NULL CHECK (reaction IN ('like', 'dislike')),
    PRIMARY KEY (target_type, target_id, visitor_id)
);
CREATE INDEX IF NOT EXISTS reactions_target_reaction_idx ON reactions (target_type, target_id, reaction);
