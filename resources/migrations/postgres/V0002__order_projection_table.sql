-- Create the order projection table
-- This table is used to store the projection of the order.
-- Used only for listing orders.
-- Serves the same purpose as a search index ala Elasticsearch.
CREATE TABLE order_projection (
    order_id VARCHAR(255) PRIMARY KEY,
    payment_status VARCHAR(255) NOT NULL,
    shipping_status VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_order_projection_created_at ON order_projection (created_at);
