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
