CREATE TABLE event (
    event_id SERIAL PRIMARY KEY,
    sequence_number BIGINT NOT NULL,
    aggregate_id VARCHAR(255) NOT NULL,
    aggregate_type VARCHAR(255) NOT NULL,
    event_type VARCHAR(255) NOT NULL,
    event_data BYTEA NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    UNIQUE (sequence_number, aggregate_id)
);

CREATE INDEX idx_event_aggregate_id ON event (aggregate_id);
