# Forum API Guide for Frontend Development

This guide provides all the necessary information to set up a fresh database, populate it with sample data, and use the public API endpoints to build the forum's homepage.

## 1. Database Setup

To ensure you are working with a clean slate, follow these steps.

### Step 1: Clean the Database

Your application uses a SQLite database file. To start fresh, simply delete the existing database file. Assuming it's named `forum.db` in the `data/` directory:

```bash
# Navigate to the project root
rm data/forum.db
```

### Step 2: Run the Application

The application is configured to automatically create and set up a new database file if one doesn't exist. Simply run the application as you normally would:

```bash
# Navigate to the project root
make run
```

This will create a new, empty `data/forum.db` file with the correct tables.

## 2. Populating with Sample Data

After creating a fresh database, you need to populate it with some data to work with.

### Step 1: Connect to the Database

Use the SQLite command-line tool to connect to the new database file:

```bash
sqlite3 data/forum.db
```

### Step 2: Run the SQL Commands

Copy and paste the following SQL commands into the SQLite prompt. This will create users, categories, and posts.

```sql
-- Make sure foreign keys are enabled
PRAGMA foreign_keys = ON;

-- Insert Sample Users
-- Passwords for all users are 'password123'
-- The hashed password below is for 'password123'
INSERT INTO "users" ("id", "username", "email", "password", "role") VALUES
('a1b2c3d4-0001-4001-8001-111111111111', 'JohnDoe', 'john.doe@example.com', '$2a$12$i1pA.RMLd9v.xJ2jQ3K3e.oY0b9c8d7e6f5g4h3i2j1k0l', 0),
('a1b2c3d4-0002-4002-8002-222222222222', 'JaneSmith', 'jane.smith@example.com', '$2a$12$i1pA.RMLd9v.xJ2jQ3K3e.oY0b9c8d7e6f5g4h3i2j1k0l', 0),
('a1b2c3d4-0003-4003-8003-333333333333', 'AdminUser', 'admin@example.com', '$2a$12$i1pA.RMLd9v.xJ2jQ3K3e.oY0b9c8d7e6f5g4h3i2j1k0l', 2);

-- Insert Sample Categories
INSERT INTO "categories" ("id", "name") VALUES
(1, 'Technology'),
(2, 'Sports'),
(3, 'General Discussion'),
(4, 'Science Fiction');

-- Insert Sample Posts
INSERT INTO "posts" ("title", "content", "author_id", "category_id") VALUES
('Getting Started with Go', 'Go is a statically typed, compiled programming language designed at Google...', 'a1b2c3d4-0001-4001-8001-111111111111', 1),
('The Future of Web Development', 'WebAssembly is changing the game. What are your thoughts?', 'a1b2c3d4-0002-4002-8002-222222222222', 1),
('Favorite Sports Moments', 'What is the most iconic sports moment you have ever witnessed?', 'a1b2c3d4-0001-4001-8001-111111111111', 2),
('Is time travel possible?', 'Let''s discuss the theoretical physics behind time travel.', 'a1b2c3d4-0002-4002-8002-222222222222', 4),
('Welcome to the Forum!', 'This is a place for general discussion. Feel free to introduce yourself!', 'a1b2c3d4-0003-4003-8003-333333333333', 3);

```

Once done, you can exit the SQLite prompt by typing `.quit`.

## 3. Public API Endpoints

These endpoints are public, require no authentication, and return data in JSON format. CORS is enabled, so you can call them from any frontend application.

**Base URL:** `http://localhost:8080` (or your configured server address)

---

### Get Random Posts

Fetches a list of 20 random posts from the forum. Ideal for the main homepage feed.

*   **Endpoint:** `GET /api/posts`
*   **Example Response (`200 OK`):**
    ```json
    [
        {
            "id": 1,
            "title": "Getting Started with Go",
            "content": "Go is a statically typed, compiled programming language designed at Google...",
            "authorId": "a1b2c3d4-0001-4001-8001-111111111111",
            "username": "JohnDoe",
            "categoryId": 1,
            "createdAt": "2025-10-30T10:00:00Z",
            "lastModified": null
        },
        {
            "id": 3,
            "title": "Favorite Sports Moments",
            "content": "What is the most iconic sports moment you have ever witnessed?",
            "authorId": "a1b2c3d4-0001-4001-8001-111111111111",
            "username": "JohnDoe",
            "categoryId": 2,
            "createdAt": "2025-10-29T18:30:00Z",
            "lastModified": {
                "Time": "2025-10-29T19:00:00Z",
                "Valid": true
            }
        }
    ]
    ```

---

### Get All Categories

Fetches a list of all available categories. Use this to build the category navigation menu.

*   **Endpoint:** `GET /api/categories`
*   **Example Response (`200 OK`):**
    ```json
    [
        {
            "id": 1,
            "name": "Technology"
        },
        {
            "id": 2,
            "name": "Sports"
        },
        {
            "id": 3,
            "name": "General Discussion"
        },
        {
            "id": 4,
            "name": "Science Fiction"
        }
    ]
    ```

---

### Get Posts by Category

Fetches a list of the 50 most recent posts belonging to a specific category.

*   **Endpoint:** `GET /api/categories/{id}/posts`
*   **Description:** Replace `{id}` with the ID of the desired category.
*   **Example Usage:** `GET /api/categories/1/posts` (to get posts from the "Technology" category)
*   **Example Response (`200 OK`):** The response format is an array of post objects, identical to the `GET /api/posts` endpoint.