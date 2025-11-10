-- +goose Up
-- +goose StatementBegin
CREATE TABLE subscriptions (
  id UUID PRIMARY KEY DEFAULT uuidv7 (),
  service_name VARCHAR(255) NOT NULL,
  monthly_cost INTEGER NOT NULL,
  user_id UUID NOT NULL,
  start_date DATE NOT NULL,
  end_date DATE,
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_subscriptions_user_id ON subscriptions (user_id);

CREATE INDEX idx_subscriptions_service_name ON subscriptions (service_name);

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS subscriptions;

-- +goose StatementEnd
