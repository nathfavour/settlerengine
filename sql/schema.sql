CREATE TABLE invoices (
    id VARCHAR(255) PRIMARY KEY,
    amount VARCHAR(255) NOT NULL,
    currency VARCHAR(50) NOT NULL,
    status VARCHAR(50) NOT NULL,
    created_at BIGINT NOT NULL,
    expires_at BIGINT NOT NULL
);

CREATE TABLE verified_payments (
    signature VARCHAR(255) PRIMARY KEY,
    signer VARCHAR(255) NOT NULL,
    amount VARCHAR(255) NOT NULL,
    asset VARCHAR(255) NOT NULL,
    nonce VARCHAR(255) NOT NULL,
    verified_at BIGINT NOT NULL
);

CREATE TABLE webhook_configs (
    id VARCHAR(255) PRIMARY KEY,
    url VARCHAR(512) NOT NULL,
    secret VARCHAR(255) NOT NULL,
    events VARCHAR(255) NOT NULL,
    created_at BIGINT NOT NULL
);

CREATE TABLE webhook_deliveries (
    id VARCHAR(255) PRIMARY KEY,
    config_id VARCHAR(255) NOT NULL,
    payload TEXT NOT NULL,
    event VARCHAR(100) NOT NULL,
    status VARCHAR(50) NOT NULL,
    attempts INT NOT NULL,
    next_attempt_at BIGINT NOT NULL,
    created_at BIGINT NOT NULL
);

CREATE TABLE client_reputations (
    client_address VARCHAR(255) PRIMARY KEY,
    score INT NOT NULL,
    total_payments VARCHAR(255) NOT NULL,
    last_payment_at BIGINT NOT NULL
);

CREATE TABLE pricing_policies (
    resource_path VARCHAR(255) PRIMARY KEY,
    base_price VARCHAR(255) NOT NULL,
    currency VARCHAR(50) NOT NULL,
    surge_multiplier DOUBLE PRECISION NOT NULL
);
