-- +goose Up
ALTER TABLE deep_analysis_usage_counters
    DROP CONSTRAINT chk_deep_analysis_usage_limit;

UPDATE deep_analysis_usage_counters
SET limit_count = 100;

ALTER TABLE deep_analysis_usage_counters
    ADD CONSTRAINT chk_deep_analysis_usage_limit
        CHECK (limit_count = 100);

-- +goose Down
ALTER TABLE deep_analysis_usage_counters
    DROP CONSTRAINT chk_deep_analysis_usage_limit;

UPDATE deep_analysis_usage_counters
SET used_count = LEAST(used_count, 2),
    limit_count = 2;

ALTER TABLE deep_analysis_usage_counters
    ADD CONSTRAINT chk_deep_analysis_usage_limit
        CHECK (limit_count = 2);
