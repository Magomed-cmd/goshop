-- DROP всё старое, если есть
DROP TABLE IF EXISTS reviews, order_items, orders, cart_items, carts, product_categories, products, categories, user_addresses, users, roles CASCADE;
DROP TYPE IF EXISTS order_status;

-- UUID extension можно оставить на будущее
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE roles (
                       id      BIGSERIAL PRIMARY KEY,
                       uuid    UUID NOT NULL UNIQUE,
                       name    TEXT NOT NULL UNIQUE
);

CREATE TABLE users (
                       id              BIGSERIAL PRIMARY KEY,
                       uuid            UUID NOT NULL UNIQUE,
                       email           TEXT NOT NULL UNIQUE,
                       password_hash   TEXT NOT NULL,
                       name            TEXT,
                       phone           TEXT,
                       role_id         BIGINT REFERENCES roles(id) ON DELETE SET NULL,
                       created_at      TIMESTAMP NOT NULL DEFAULT now()
);

CREATE TABLE user_addresses (
                                id          BIGSERIAL PRIMARY KEY,
                                uuid        UUID NOT NULL UNIQUE,
                                user_id     BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                                address     TEXT NOT NULL,
                                city        TEXT,
                                postal_code TEXT,
                                country     TEXT,
                                created_at  TIMESTAMP NOT NULL DEFAULT now()
);

CREATE TABLE categories (
                            id          BIGSERIAL PRIMARY KEY,
                            uuid        UUID NOT NULL UNIQUE,
                            name        TEXT NOT NULL,
                            description TEXT
);

CREATE TABLE products (
                          id          BIGSERIAL PRIMARY KEY,
                          uuid        UUID NOT NULL UNIQUE,
                          name        TEXT NOT NULL,
                          description TEXT,
                          price       NUMERIC(12,2) NOT NULL,
                          stock       INT NOT NULL DEFAULT 0,
                          created_at  TIMESTAMP NOT NULL DEFAULT now(),
                          updated_at  TIMESTAMP NOT NULL DEFAULT now()
);

CREATE TABLE product_categories (
                                    product_id  BIGINT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
                                    category_id BIGINT NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
                                    PRIMARY KEY (product_id, category_id)
);

CREATE TABLE carts (
                       id          BIGSERIAL PRIMARY KEY,
                       uuid        UUID NOT NULL UNIQUE,
                       user_id     BIGINT NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
                       created_at  TIMESTAMP NOT NULL DEFAULT now()
);

CREATE TABLE cart_items (
                            cart_id     BIGINT NOT NULL REFERENCES carts(id) ON DELETE CASCADE,
                            product_id  BIGINT NOT NULL REFERENCES products(id),
                            quantity    INT NOT NULL CHECK (quantity > 0),
                            PRIMARY KEY (cart_id, product_id)
);

DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'order_status') THEN
        CREATE TYPE order_status AS ENUM ('pending', 'paid', 'shipped', 'delivered', 'cancelled');
    END IF;
END $$;

CREATE TABLE orders (
                        id              BIGSERIAL PRIMARY KEY,
                        uuid            UUID NOT NULL UNIQUE,
                        user_id         BIGINT NOT NULL REFERENCES users(id) ON DELETE SET NULL,
                        address_id      BIGINT REFERENCES user_addresses(id),
                        total_price     NUMERIC(12,2) NOT NULL,
                        status          order_status NOT NULL DEFAULT 'pending',
                        created_at      TIMESTAMP NOT NULL DEFAULT now(),
                        updated_at      TIMESTAMP NOT NULL DEFAULT now()
);

CREATE TABLE order_items (
                             order_id        BIGINT NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
                             product_id      BIGINT NOT NULL REFERENCES products(id),
                             product_name    TEXT NOT NULL,
                             price_at_order  NUMERIC(12,2) NOT NULL,
                             quantity        INT NOT NULL CHECK (quantity > 0),
                             PRIMARY KEY (order_id, product_id)
);

CREATE INDEX idx_products_name ON products (name);
CREATE INDEX idx_products_price ON products (price);

CREATE TABLE reviews (
                         id          BIGSERIAL PRIMARY KEY,
                         uuid        UUID NOT NULL UNIQUE,
                         product_id  BIGINT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
                         user_id     BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                         rating      INT NOT NULL CHECK (rating >= 1 AND rating <= 5),
                         comment     TEXT,
                         created_at  TIMESTAMP NOT NULL DEFAULT now(),
                         UNIQUE (product_id, user_id)
);