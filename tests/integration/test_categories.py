# tests/integration/test_categories.py
import pytest

class TestCategories:
    def test_get_all_categories_empty(self, api_client):
        """Тест получения пустого списка категорий"""
        response = api_client.get("/categories")

        assert response.status_code == 200
        data = response.json()
        assert isinstance(data, list)

    def test_get_all_categories_with_data(self, api_client, admin_token):
        """Тест получения списка категорий с данными"""
        if not admin_token:
            pytest.skip("Admin token not available")

        # Создаем категорию
        category_data = {
            "name": "Test Category for List",
            "description": "Test description"
        }
        headers = {"Authorization": f"Bearer {admin_token}"}
        create_response = api_client.post("/admin/categories", json=category_data, headers=headers)
        assert create_response.status_code == 201

        # Получаем список категорий
        response = api_client.get("/categories")
        assert response.status_code == 200

        data = response.json()
        assert isinstance(data, list)
        assert len(data) >= 1

        # Проверяем структуру категории
        category = data[0]
        assert "id" in category
        assert "uuid" in category
        assert "name" in category
        assert "description" in category
        assert "product_count" in category

    def test_get_category_by_id_success(self, api_client, admin_token):
        """Тест успешного получения категории по ID"""
        if not admin_token:
            pytest.skip("Admin token not available")

        # Создаем категорию
        category_data = {
            "name": "Category for Get by ID",
            "description": "Test description for get"
        }
        headers = {"Authorization": f"Bearer {admin_token}"}
        create_response = api_client.post("/admin/categories", json=category_data, headers=headers)
        assert create_response.status_code == 201

        category_id = create_response.json()["id"]

        # Получаем категорию по ID
        response = api_client.get(f"/categories/{category_id}")
        assert response.status_code == 200

        data = response.json()
        assert data["id"] == category_id
        assert data["name"] == "Category for Get by ID"
        assert data["description"] == "Test description for get"

    def test_get_category_by_id_not_found(self, api_client):
        """Тест получения несуществующей категории"""
        response = api_client.get("/categories/99999")
        assert response.status_code == 404

    def test_get_category_by_id_invalid_id(self, api_client):
        """Тест получения категории с невалидным ID"""
        response = api_client.get("/categories/invalid")
        assert response.status_code == 400

    def test_create_category_success(self, api_client, admin_token):
        """Тест успешного создания категории"""
        if not admin_token:
            pytest.skip("Admin token not available")

        category_data = {
            "name": "New Test Category",
            "description": "New test description"
        }

        headers = {"Authorization": f"Bearer {admin_token}"}
        response = api_client.post("/admin/categories", json=category_data, headers=headers)

        assert response.status_code == 201
        data = response.json()
        assert data["name"] == "New Test Category"
        assert data["description"] == "New test description"
        assert data["product_count"] == 0
        assert "id" in data
        assert "uuid" in data

    def test_create_category_unauthorized(self, api_client):
        """Тест создания категории без авторизации"""
        category_data = {
            "name": "Unauthorized Category",
            "description": "Should fail"
        }

        response = api_client.post("/admin/categories", json=category_data)
        assert response.status_code == 401

    def test_create_category_invalid_data(self, api_client, admin_token):
        """Тест создания категории с невалидными данными"""
        if not admin_token:
            pytest.skip("Admin token not available")

        # Пустое имя
        category_data = {
            "name": "",
            "description": "Valid description"
        }

        headers = {"Authorization": f"Bearer {admin_token}"}
        response = api_client.post("/admin/categories", json=category_data, headers=headers)
        assert response.status_code == 400

    def test_create_category_missing_name(self, api_client, admin_token):
        """Тест создания категории без имени"""
        if not admin_token:
            pytest.skip("Admin token not available")

        category_data = {
            "description": "Valid description"
        }

        headers = {"Authorization": f"Bearer {admin_token}"}
        response = api_client.post("/admin/categories", json=category_data, headers=headers)
        assert response.status_code == 400

    def test_update_category_success(self, api_client, admin_token):
        """Тест успешного обновления категории"""
        if not admin_token:
            pytest.skip("Admin token not available")

        # Создаем категорию
        category_data = {
            "name": "Category to Update",
            "description": "Original description"
        }
        headers = {"Authorization": f"Bearer {admin_token}"}
        create_response = api_client.post("/admin/categories", json=category_data, headers=headers)
        assert create_response.status_code == 201

        category_id = create_response.json()["id"]

        # Обновляем категорию
        update_data = {
            "name": "Updated Category Name",
            "description": "Updated description"
        }

        response = api_client.put(f"/admin/categories/{category_id}", json=update_data, headers=headers)
        assert response.status_code == 200

        data = response.json()
        assert data["name"] == "Updated Category Name"
        assert data["description"] == "Updated description"

    def test_update_category_not_found(self, api_client, admin_token):
        """Тест обновления несуществующей категории"""
        if not admin_token:
            pytest.skip("Admin token not available")

        update_data = {
            "name": "Updated Name",
            "description": "Updated description"
        }
        headers = {"Authorization": f"Bearer {admin_token}"}

        response = api_client.put("/admin/categories/99999", json=update_data, headers=headers)
        assert response.status_code == 404

    def test_update_category_unauthorized(self, api_client):
        """Тест обновления категории без авторизации"""
        update_data = {
            "name": "Updated Name"
        }

        response = api_client.put("/admin/categories/1", json=update_data)
        assert response.status_code == 401

    def test_delete_category_success(self, api_client, admin_token):
        """Тест успешного удаления категории"""
        if not admin_token:
            pytest.skip("Admin token not available")

        # Создаем категорию
        category_data = {
            "name": "Category to Delete",
            "description": "Will be deleted"
        }
        headers = {"Authorization": f"Bearer {admin_token}"}
        create_response = api_client.post("/admin/categories", json=category_data, headers=headers)
        assert create_response.status_code == 201

        category_id = create_response.json()["id"]

        # Удаляем категорию
        response = api_client.delete(f"/admin/categories/{category_id}", headers=headers)
        assert response.status_code == 204

        # Проверяем что категория удалена
        get_response = api_client.get(f"/categories/{category_id}")
        assert get_response.status_code == 404

    def test_delete_category_not_found(self, api_client, admin_token):
        """Тест удаления несуществующей категории"""
        if not admin_token:
            pytest.skip("Admin token not available")

        headers = {"Authorization": f"Bearer {admin_token}"}
        response = api_client.delete("/admin/categories/99999", headers=headers)
        assert response.status_code == 404

    def test_delete_category_unauthorized(self, api_client):
        """Тест удаления категории без авторизации"""
        response = api_client.delete("/admin/categories/1")
        assert response.status_code == 401

    def test_delete_category_invalid_id(self, api_client, admin_token):
        """Тест удаления категории с невалидным ID"""
        if not admin_token:
            pytest.skip("Admin token not available")

        headers = {"Authorization": f"Bearer {admin_token}"}
        response = api_client.delete("/admin/categories/invalid", headers=headers)
        assert response.status_code == 400