CREATE TABLE users (
  id VARCHAR(255) PRIMARY KEY,
  name VARCHAR(255) NOT NULL,
  age INT NOT NULL
);
CREATE TABLE books (id VARCHAR(255) PRIMARY KEY, name VARCHAR(255) NOT NULL);
CREATE TABLE lending (
  user_id VARCHAR(255) DEFAULT NULL,
  book_id VARCHAR(255) DEFAULT NULL
);
