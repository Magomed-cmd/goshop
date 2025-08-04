# tests/integration/test_auth.py
import pytest

class TestAuth:
    def test_register_success(self, api_client, unique_email):
        """Тест успешной регистрации"""
        user_data = {
            "email": unique_email,
            "password": "password123",
            "name": "New User"
        }

        response = api_client.post("/auth/register", json=user_data)

        assert response.status_code == 201
        data = response.json()
        assert "token" in data
        assert "user" in data
        assert data["user"]["email"] == unique_email
        assert data["user"]["name"] == "New User"
        assert data["user"]["role"] == "user"

    def test_register_duplicate_email(self, api_client, test_user):
        """Тест регистрации с существующим email"""
        # Регистрируем пользователя первый раз
        api_client.post("/auth/register", json=test_user)

        # Пытаемся зарегистрировать с тем же email
        response = api_client.post("/auth/register", json=test_user)

        assert response.status_code == 409
        assert "error" in response.json()

    def test_register_invalid_email(self, api_client):
        """Тест регистрации с невалидным email"""
        user_data = {
            "email": "invalid-email",
            "password": "password123",
            "name": "Test User"
        }

        response = api_client.post("/auth/register", json=user_data)
        assert response.status_code == 400

    def test_register_short_password(self, api_client, unique_email):
        """Тест регистрации с коротким паролем"""
        user_data = {
            "email": unique_email,
            "password": "123",
            "name": "Test User"
        }

        response = api_client.post("/auth/register", json=user_data)
        assert response.status_code == 400

    def test_login_success(self, api_client, test_user):
        """Тест успешного логина"""
        # Сначала регистрируем пользователя
        api_client.post("/auth/register", json=test_user)

        # Логинимся
        login_data = {
            "email": test_user["email"],
            "password": test_user["password"]
        }
        response = api_client.post("/auth/login", json=login_data)

        assert response.status_code == 200
        data = response.json()
        assert "token" in data
        assert "user" in data
        assert data["user"]["email"] == test_user["email"]

    def test_login_invalid_credentials(self, api_client):
        """Тест логина с неверными данными"""
        login_data = {
            "email": "nonexistent@example.com",
            "password": "wrongpassword"
        }

        response = api_client.post("/auth/login", json=login_data)
        assert response.status_code == 401

    def test_login_missing_fields(self, api_client):
        """Тест логина с отсутствующими полями"""
        response = api_client.post("/auth/login", json={"email": "test@example.com"})
        assert response.status_code == 400

    def test_get_profile(self, authenticated_client):
        """Тест получения профиля"""
        response = authenticated_client.get("/api/v1/profile")

        assert response.status_code == 200
        data = response.json()
        assert "email" in data
        assert "name" in data
        assert "role" in data

    def test_get_profile_unauthorized(self, api_client):
        """Тест получения профиля без авторизации"""
        # Создаем чистый клиент без токенов
        clean_response = api_client.get("/api/v1/profile")
        assert clean_response.status_code == 401

    def test_update_profile(self, authenticated_client):
        """Тест обновления профиля"""
        update_data = {
            "name": "Updated Name",
            "phone": "+1234567890"
        }

        response = authenticated_client.put("/api/v1/profile", json=update_data)
        assert response.status_code == 200

    def test_update_profile_empty(self, authenticated_client):
        """Тест обновления профиля без данных"""
        response = authenticated_client.put("/api/v1/profile", json={})
        assert response.status_code == 400