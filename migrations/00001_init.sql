-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS users (
                                     id UUID PRIMARY KEY,
                                     user_role TEXT NOT NULL,
                                     first_name TEXT NOT NULL,
                                     last_name TEXT NOT NULL,
                                     position TEXT NOT NULL,
                                     email TEXT UNIQUE NOT NULL,
                                     hashed_password TEXT NOT NULL,
                                     is_active BOOLEAN NOT NULL
);

CREATE TABLE IF NOT EXISTS forms (
                                     id UUID PRIMARY KEY,
                                     user_id UUID NOT NULL,
                                     title TEXT NOT NULL,
                                     description TEXT,
                                     start_date TIMESTAMPTZ,
                                     end_date TIMESTAMPTZ,
                                     created_at TIMESTAMPTZ NOT NULL,
                                     reviewed_at TIMESTAMPTZ,
                                     status TEXT NOT NULL,
                                     comment TEXT,
                                     executor_id UUID,
                                     CONSTRAINT fk_forms_user FOREIGN KEY (user_id) REFERENCES users(id),
                                     CONSTRAINT fk_forms_executor FOREIGN KEY (executor_id) REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS documents (
                                        id UUID PRIMARY KEY,
                                        form_id UUID NOT NULL,
                                        object_key TEXT NOT NULL UNIQUE,
                                        original_name TEXT NOT NULL,
                                        uploaded_at TIMESTAMP NOT NULL,
                                        type TEXT NOT NULL,
                                        CONSTRAINT fk_documents_form FOREIGN KEY (form_id) REFERENCES forms(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_forms_executor_id ON forms(executor_id);
CREATE INDEX IF NOT EXISTS idx_forms_status ON forms(status);
CREATE INDEX IF NOT EXISTS idx_forms_user_id ON forms(user_id);
CREATE INDEX IF NOT EXISTS idx_documents_form_id ON documents(form_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS documents;
DROP TABLE IF EXISTS forms;
DROP TABLE IF EXISTS users;
-- +goose StatementEnd
