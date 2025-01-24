-- $ sudo -u postgres createuser mqtt2tsdb -P
-- $ sudo -u postgres createdb -O mqtt2tsdb --lc-collate C --template template0 mqtt2tsdb

CREATE TABLE metrics (
    id BIGSERIAL PRIMARY KEY ,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    metric_name TEXT NOT NULL,
    metric_value BOOLEAN
);