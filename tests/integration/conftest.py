# tests/integration/conftest.py
import pytest
import requests
import time
import random
import string

class APIClient:
    def __init__(self, base_url="http://localhost:8080"):
        self.base_url = base_url
        self.session = requests.Session()

    def get(self, endpoint, headers=None, params=None):
        return self.session.get(f"{self.base_url}{endpoint}", headers=headers, params=params)

    def post(self, endpoint, json=None, headers=None):
        return self.session.post(f"{self.base_url}{endpoint}", json=json, headers=headers)

    def put(self, endpoint, json=None, headers=None):
        return self.session.put(f"{self.base_url}{endpoint}", json=json, headers=headers)

    def delete(self, endpoint, headers=None):
        return self.session.delete(f"{self.base_url}{endpoint}", headers=headers)

@pytest.fixture  # ← УБРАЛ scope="session"!
def api_client():
    """Создает НОВЫЙ неавторизованный клиент для каждого теста"""
    return APIClient()

@pytest.fixture
def unique_email():
    """Генерирует уникальный email для каждого теста"""
    timestamp = int(time.time())
    random_suffix = ''.join(random.choices(string.ascii_lowercase, k=4))
    return f"test{timestamp}{random_suffix}@example.com"

@pytest.fixture
def test_user(unique_email):
    """Тестовый пользователь с уникальным email"""
    return {
        "email": unique_email,
        "password": "testpassword123",
        "name": "Test User"
    }

@pytest.fixture(scope="session")
def admin_token():
    """Получает admin JWT token"""
    client = APIClient()  # Отдельный клиент для админа
    admin_data = {
        "email": "admin@example.com",
        "password": "admin123"
    }

    response = client.post("/auth/login", json=admin_data)
    if response.status_code == 200:
        return response.json()["token"]
    else:
        # Если админа нет, создаем обычного пользователя и возвращаем его токен
        # В реальном проекте лучше создать админа через SQL
        return None

@pytest.fixture(scope="session")
def test_category(admin_token):
    """Создает тестовую категорию"""
    if not admin_token:
        pytest.skip("Admin token not available")

    client = APIClient()  # Отдельный клиент для создания категории
    category_data = {
        "name": "Test Electronics",
        "description": "Test category for integration tests"
    }

    headers = {"Authorization": f"Bearer {admin_token}"}
    response = client.post("/admin/categories", json=category_data, headers=headers)

    if response.status_code == 201:
        return response.json()
    else:
        pytest.skip("Failed to create test category")

@pytest.fixture
def authenticated_client(test_user):
    """Клиент с авторизованным пользователем"""
    client = APIClient()  # НОВЫЙ клиент, не модифицируем общий!

    # Регистрируем пользователя
    register_response = client.post("/auth/register", json=test_user)

    if register_response.status_code == 201:
        token = register_response.json()["token"]
    elif register_response.status_code == 409:
        # Пользователь уже существует, логинимся
        login_response = client.post("/auth/login", json={
            "email": test_user["email"],
            "password": test_user["password"]
        })
        if login_response.status_code == 200:
            token = login_response.json()["token"]
        else:
            pytest.fail("Failed to login existing user")
    else:
        pytest.fail(f"Registration failed: {register_response.status_code}")

    # Добавляем токен в заголовки ЭТОГО клиента
    client.session.headers.update({"Authorization": f"Bearer {token}"})
    return client