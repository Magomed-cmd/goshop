# tests/integration/test_products.py
import pytest

class TestProducts:
    def test_get_products_empty(self, api_client):
        """Тест получения пустого списка продуктов"""
        response = api_client.get("/products")

        assert response.status_code == 200
        data = response.json()
        assert "products" in data
        assert "total" in data
        assert data["total"] >= 0

    def test_get_products_with_pagination(self, api_client):
        """Тест пагинации продуктов"""
        response = api_client.get("/products", params={
            "page": 1,
            "limit": 5
        })

        assert response.status_code == 200
        data = response.json()
        assert "products" in data
        assert "page" in data
        assert "limit" in data
        assert data["page"] == 1
        assert data["limit"] == 5

    def test_create_product_success(self, api_client, admin_token, test_category):
        """Тест успешного создания продукта"""
        if not admin_token:
            pytest.skip("Admin token not available")

        product_data = {
            "name": "Test Product",
            "description": "Test Description",
            "price": "99.99",
            "stock": 10,
            "category_ids": [test_category["id"]]
        }

        headers = {"Authorization": f"Bearer {admin_token}"}
        response = api_client.post("/admin/products", json=product_data, headers=headers)

        assert response.status_code == 201
        data = response.json()
        assert data["name"] == "Test Product"
        assert data["price"] == "99.99"
        assert data["stock"] == 10
        assert len(data["categories"]) == 1

    def test_create_product_unauthorized(self, api_client):
        """Тест создания продукта без авторизации"""
        product_data = {
            "name": "Test Product",
            "price": "99.99",
            "stock": 10,
            "category_ids": [1]
        }

        response = api_client.post("/admin/products", json=product_data)
        assert response.status_code == 401

    def test_create_product_invalid_data(self, api_client, admin_token):
        """Тест создания продукта с невалидными данными"""
        if not admin_token:
            pytest.skip("Admin token not available")

        product_data = {
            "name": "",  # Пустое имя
            "price": "invalid",  # Невалидная цена
            "stock": -1  # Отрицательный stock
        }

        headers = {"Authorization": f"Bearer {admin_token}"}
        response = api_client.post("/admin/products", json=product_data, headers=headers)
        assert response.status_code == 400

    def test_get_product_by_id(self, api_client, admin_token, test_category):
        """Тест получения продукта по ID"""
        if not admin_token:
            pytest.skip("Admin token not available")

        # Создаем продукт
        product_data = {
            "name": "Test Product for Get",
            "description": "Test Description",
            "price": "99.99",
            "stock": 10,
            "category_ids": [test_category["id"]]
        }

        headers = {"Authorization": f"Bearer {admin_token}"}
        create_response = api_client.post("/admin/products", json=product_data, headers=headers)
        assert create_response.status_code == 201

        product_id = create_response.json()["id"]

        # Получаем продукт по ID
        response = api_client.get(f"/products/{product_id}")
        assert response.status_code == 200

        data = response.json()
        assert data["id"] == product_id
        assert data["name"] == "Test Product for Get"

    def test_get_product_not_found(self, api_client):
        """Тест получения несуществующего продукта"""
        response = api_client.get("/products/99999")
        assert response.status_code == 404

    def test_get_product_invalid_id(self, api_client):
        """Тест получения продукта с невалидным ID"""
        response = api_client.get("/products/invalid")
        assert response.status_code == 400

    def test_update_product(self, api_client, admin_token, test_category):
        """Тест обновления продукта"""
        if not admin_token:
            pytest.skip("Admin token not available")

        # Создаем продукт
        product_data = {
            "name": "Product to Update",
            "price": "50.00",
            "stock": 20,
            "category_ids": [test_category["id"]]
        }

        headers = {"Authorization": f"Bearer {admin_token}"}
        create_response = api_client.post("/admin/products", json=product_data, headers=headers)
        product_id = create_response.json()["id"]

        # Обновляем продукт
        update_data = {
            "name": "Updated Product",
            "price": "75.00"
        }

        response = api_client.put(f"/admin/products/{product_id}", json=update_data, headers=headers)
        assert response.status_code == 200

        data = response.json()
        assert data["name"] == "Updated Product"
        assert data["price"] == "75.00"

    def test_update_product_not_found(self, api_client, admin_token):
        """Тест обновления несуществующего продукта"""
        if not admin_token:
            pytest.skip("Admin token not available")

        update_data = {"name": "Updated Name"}
        headers = {"Authorization": f"Bearer {admin_token}"}

        response = api_client.put("/admin/products/99999", json=update_data, headers=headers)
        assert response.status_code == 404

    def test_delete_product(self, api_client, admin_token, test_category):
        """Тест удаления продукта"""
        if not admin_token:
            pytest.skip("Admin token not available")

        # Создаем продукт
        product_data = {
            "name": "Product to Delete",
            "price": "30.00",
            "stock": 5,
            "category_ids": [test_category["id"]]
        }

        headers = {"Authorization": f"Bearer {admin_token}"}
        create_response = api_client.post("/admin/products", json=product_data, headers=headers)
        product_id = create_response.json()["id"]

        # Удаляем продукт
        response = api_client.delete(f"/admin/products/{product_id}", headers=headers)
        assert response.status_code == 204

        # Проверяем что продукт удален
        get_response = api_client.get(f"/products/{product_id}")
        assert get_response.status_code == 404

    def test_get_products_with_filters(self, api_client):
        """Тест получения продуктов с фильтрами"""
        response = api_client.get("/products", params={
            "min_price": "10.00",
            "max_price": "100.00",
            "sort_by": "price",
            "sort_order": "asc"
        })

        assert response.status_code == 200
        data = response.json()
        assert "products" in data

    def test_get_products_by_category(self, api_client, test_category):
        """Тест получения продуктов по категории"""
        response = api_client.get(f"/products/category/{test_category['id']}")

        assert response.status_code == 200
        data = response.json()
        assert "products" in data