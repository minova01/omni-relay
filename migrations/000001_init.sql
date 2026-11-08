BEGIN;

-- Omni Relay initial PostgreSQL schema.
-- Business timestamps are Unix milliseconds stored as BIGINT.
-- Relationships are maintained by the application; this schema intentionally has no foreign keys.

CREATE TABLE users (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    username VARCHAR(32) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role SMALLINT NOT NULL DEFAULT 1,
    status SMALLINT NOT NULL DEFAULT 1,
    quota BIGINT NOT NULL DEFAULT 0,
    used_quota BIGINT NOT NULL DEFAULT 0,
    request_count BIGINT NOT NULL DEFAULT 0,
    created_at BIGINT NOT NULL,
    updated_at BIGINT NOT NULL,
    deleted_at BIGINT NOT NULL DEFAULT 0,
    CONSTRAINT users_username_not_blank CHECK (length(btrim(username)) > 0),
    CONSTRAINT users_role_valid CHECK (role IN (1, 10)),
    CONSTRAINT users_status_valid CHECK (status IN (1, 2)),
    CONSTRAINT users_quota_nonnegative CHECK (quota >= 0),
    CONSTRAINT users_used_quota_nonnegative CHECK (used_quota >= 0),
    CONSTRAINT users_request_count_nonnegative CHECK (request_count >= 0),
    CONSTRAINT users_timestamps_nonnegative CHECK (
        created_at >= 0 AND updated_at >= 0 AND deleted_at >= 0
    )
);

CREATE UNIQUE INDEX uk_users_username_active
    ON users (username)
    WHERE deleted_at = 0;

COMMENT ON COLUMN users.role IS '1=user, 10=admin';
COMMENT ON COLUMN users.status IS '1=enabled, 2=disabled';
COMMENT ON COLUMN users.quota IS 'Current balance in integer internal quota units';
COMMENT ON COLUMN users.used_quota IS 'Lifetime consumed integer quota units';
COMMENT ON COLUMN users.request_count IS 'Lifetime settled request count';

CREATE TABLE api_keys (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id BIGINT NOT NULL,
    name VARCHAR(64) NOT NULL,
    key_prefix VARCHAR(32) NOT NULL,
    key_hash CHAR(64) NOT NULL,
    key_hint VARCHAR(8) NOT NULL,
    status SMALLINT NOT NULL DEFAULT 1,
    used_quota BIGINT NOT NULL DEFAULT 0,
    request_count BIGINT NOT NULL DEFAULT 0,
    last_used_at BIGINT NOT NULL DEFAULT 0,
    created_at BIGINT NOT NULL,
    updated_at BIGINT NOT NULL,
    deleted_at BIGINT NOT NULL DEFAULT 0,
    CONSTRAINT api_keys_name_not_blank CHECK (length(btrim(name)) > 0),
    CONSTRAINT api_keys_prefix_not_blank CHECK (length(btrim(key_prefix)) > 0),
    CONSTRAINT api_keys_hash_hex CHECK (key_hash ~ '^[0-9a-f]{64}$'),
    CONSTRAINT api_keys_status_valid CHECK (status IN (1, 2)),
    CONSTRAINT api_keys_usage_nonnegative CHECK (used_quota >= 0 AND request_count >= 0),
    CONSTRAINT api_keys_timestamps_nonnegative CHECK (
        last_used_at >= 0 AND created_at >= 0 AND updated_at >= 0 AND deleted_at >= 0
    )
);

CREATE UNIQUE INDEX uk_api_keys_prefix_active
    ON api_keys (key_prefix)
    WHERE deleted_at = 0;

CREATE INDEX idx_api_keys_user_created
    ON api_keys (user_id, created_at DESC, id DESC)
    WHERE deleted_at = 0;

COMMENT ON COLUMN api_keys.key_prefix IS 'Public lookup prefix; the complete API key is never persisted';
COMMENT ON COLUMN api_keys.key_hash IS 'Lowercase hex SHA-256 digest of the secret portion';
COMMENT ON COLUMN api_keys.status IS '1=enabled, 2=disabled';
COMMENT ON COLUMN api_keys.used_quota IS 'Lifetime consumed integer quota units';

CREATE TABLE models (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    code VARCHAR(128) NOT NULL,
    display_name VARCHAR(128) NOT NULL,
    status SMALLINT NOT NULL DEFAULT 1,
    input_price BIGINT NOT NULL DEFAULT 0,
    output_price BIGINT NOT NULL DEFAULT 0,
    created_at BIGINT NOT NULL,
    updated_at BIGINT NOT NULL,
    deleted_at BIGINT NOT NULL DEFAULT 0,
    CONSTRAINT models_code_not_blank CHECK (length(btrim(code)) > 0),
    CONSTRAINT models_display_name_not_blank CHECK (length(btrim(display_name)) > 0),
    CONSTRAINT models_status_valid CHECK (status IN (1, 2)),
    CONSTRAINT models_prices_nonnegative CHECK (input_price >= 0 AND output_price >= 0),
    CONSTRAINT models_timestamps_nonnegative CHECK (
        created_at >= 0 AND updated_at >= 0 AND deleted_at >= 0
    )
);

CREATE UNIQUE INDEX uk_models_code_active
    ON models (code)
    WHERE deleted_at = 0;

CREATE INDEX idx_models_status
    ON models (status, id)
    WHERE deleted_at = 0;

COMMENT ON COLUMN models.code IS 'Model name exposed to clients';
COMMENT ON COLUMN models.status IS '1=enabled, 2=disabled';
COMMENT ON COLUMN models.input_price IS 'Integer quota units per 1,000,000 input tokens';
COMMENT ON COLUMN models.output_price IS 'Integer quota units per 1,000,000 output tokens';

CREATE TABLE channels (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name VARCHAR(128) NOT NULL,
    channel_type VARCHAR(32) NOT NULL,
    base_url VARCHAR(512) NOT NULL,
    secret_ciphertext TEXT NOT NULL,
    status SMALLINT NOT NULL DEFAULT 1,
    used_quota BIGINT NOT NULL DEFAULT 0,
    request_count BIGINT NOT NULL DEFAULT 0,
    created_at BIGINT NOT NULL,
    updated_at BIGINT NOT NULL,
    deleted_at BIGINT NOT NULL DEFAULT 0,
    CONSTRAINT channels_name_not_blank CHECK (length(btrim(name)) > 0),
    CONSTRAINT channels_type_not_blank CHECK (length(btrim(channel_type)) > 0),
    CONSTRAINT channels_base_url_not_blank CHECK (length(btrim(base_url)) > 0),
    CONSTRAINT channels_secret_not_blank CHECK (length(btrim(secret_ciphertext)) > 0),
    CONSTRAINT channels_status_valid CHECK (status IN (1, 2)),
    CONSTRAINT channels_usage_nonnegative CHECK (used_quota >= 0 AND request_count >= 0),
    CONSTRAINT channels_timestamps_nonnegative CHECK (
        created_at >= 0 AND updated_at >= 0 AND deleted_at >= 0
    )
);

CREATE UNIQUE INDEX uk_channels_name_active
    ON channels (name)
    WHERE deleted_at = 0;

CREATE INDEX idx_channels_status
    ON channels (status, id)
    WHERE deleted_at = 0;

COMMENT ON COLUMN channels.channel_type IS 'Protocol adapter type; first release supports openai_compatible';
COMMENT ON COLUMN channels.secret_ciphertext IS 'Application-encrypted upstream credential';
COMMENT ON COLUMN channels.status IS '1=enabled, 2=disabled';
COMMENT ON COLUMN channels.used_quota IS 'Lifetime relayed integer quota units';

CREATE TABLE channel_models (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    channel_id BIGINT NOT NULL,
    model_id BIGINT NOT NULL,
    upstream_model VARCHAR(128) NOT NULL,
    priority INTEGER NOT NULL DEFAULT 0,
    supports_stream BOOLEAN NOT NULL DEFAULT TRUE,
    supports_tools BOOLEAN NOT NULL DEFAULT FALSE,
    status SMALLINT NOT NULL DEFAULT 1,
    created_at BIGINT NOT NULL,
    updated_at BIGINT NOT NULL,
    deleted_at BIGINT NOT NULL DEFAULT 0,
    CONSTRAINT channel_models_upstream_model_not_blank CHECK (length(btrim(upstream_model)) > 0),
    CONSTRAINT channel_models_status_valid CHECK (status IN (1, 2)),
    CONSTRAINT channel_models_timestamps_nonnegative CHECK (
        created_at >= 0 AND updated_at >= 0 AND deleted_at >= 0
    )
);

CREATE UNIQUE INDEX uk_channel_models_mapping_active
    ON channel_models (channel_id, model_id)
    WHERE deleted_at = 0;

CREATE INDEX idx_channel_models_route
    ON channel_models (model_id, status, priority DESC, id ASC)
    WHERE deleted_at = 0;

CREATE INDEX idx_channel_models_channel
    ON channel_models (channel_id, status, id)
    WHERE deleted_at = 0;

COMMENT ON COLUMN channel_models.priority IS 'Higher values are selected first';
COMMENT ON COLUMN channel_models.status IS '1=enabled, 2=disabled';

CREATE TABLE consumption_records (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    request_id VARCHAR(64) NOT NULL,
    user_id BIGINT NOT NULL,
    api_key_id BIGINT NOT NULL,
    model_id BIGINT NOT NULL,
    channel_id BIGINT NOT NULL,
    external_model VARCHAR(128) NOT NULL,
    upstream_model VARCHAR(128) NOT NULL,
    prompt_tokens BIGINT NOT NULL DEFAULT 0,
    completion_tokens BIGINT NOT NULL DEFAULT 0,
    consumed_quota BIGINT NOT NULL DEFAULT 0,
    request_status SMALLINT NOT NULL,
    aggregate_status SMALLINT NOT NULL DEFAULT 1,
    aggregate_retry_count INTEGER NOT NULL DEFAULT 0,
    aggregated_at BIGINT NOT NULL DEFAULT 0,
    is_stream BOOLEAN NOT NULL DEFAULT FALSE,
    upstream_status INTEGER NOT NULL DEFAULT 0,
    started_at BIGINT NOT NULL,
    finished_at BIGINT NOT NULL DEFAULT 0,
    created_at BIGINT NOT NULL,
    CONSTRAINT consumption_request_id_not_blank CHECK (length(btrim(request_id)) > 0),
    CONSTRAINT consumption_models_not_blank CHECK (
        length(btrim(external_model)) > 0 AND length(btrim(upstream_model)) > 0
    ),
    CONSTRAINT consumption_usage_nonnegative CHECK (
        prompt_tokens >= 0 AND completion_tokens >= 0 AND consumed_quota >= 0
    ),
    CONSTRAINT consumption_request_status_valid CHECK (request_status IN (1, 2)),
    CONSTRAINT consumption_aggregate_status_valid CHECK (aggregate_status IN (1, 2)),
    CONSTRAINT consumption_aggregate_retry_nonnegative CHECK (aggregate_retry_count >= 0),
    CONSTRAINT consumption_upstream_status_valid CHECK (upstream_status BETWEEN 0 AND 599),
    CONSTRAINT consumption_timestamps_nonnegative CHECK (
        started_at >= 0 AND finished_at >= 0 AND aggregated_at >= 0 AND created_at >= 0
    )
);

CREATE UNIQUE INDEX uk_consumption_records_request_id
    ON consumption_records (request_id);

CREATE INDEX idx_consumption_records_user_created
    ON consumption_records (user_id, created_at DESC, id DESC);

CREATE INDEX idx_consumption_records_user_model_created
    ON consumption_records (user_id, model_id, created_at DESC);

CREATE INDEX idx_consumption_records_api_key_created
    ON consumption_records (api_key_id, created_at DESC);

CREATE INDEX idx_consumption_records_channel_created
    ON consumption_records (channel_id, created_at DESC);

CREATE INDEX idx_consumption_records_pending
    ON consumption_records (id)
    WHERE aggregate_status = 1;

COMMENT ON COLUMN consumption_records.request_status IS '1=settled, 2=failed';
COMMENT ON COLUMN consumption_records.aggregate_status IS '1=pending, 2=done';
COMMENT ON COLUMN consumption_records.consumed_quota IS 'Final integer quota deducted from the user';
COMMENT ON TABLE consumption_records IS 'Business consumption data; prompts and model responses are intentionally not stored';

COMMIT;
