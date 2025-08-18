-- Таблица для аватарок пользователей (1 к 1)
CREATE TABLE user_avatars (
                              id BIGSERIAL PRIMARY KEY,
                              user_id BIGINT NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
                              image_url TEXT NOT NULL,
                              created_at TIMESTAMP NOT NULL DEFAULT now(),
                              updated_at TIMESTAMP NOT NULL DEFAULT now()
);

-- Таблица для картинок товаров (1 ко многим)
CREATE TABLE product_images (
                                id BIGSERIAL PRIMARY KEY,
                                product_id BIGINT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
                                image_url TEXT NOT NULL,
                                position INT NOT NULL DEFAULT 1, -- порядок отображения
                                created_at TIMESTAMP NOT NULL DEFAULT now(),
                                updated_at TIMESTAMP NOT NULL DEFAULT now()
);