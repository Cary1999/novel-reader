package mysql

const schemaSQL = `
CREATE TABLE IF NOT EXISTS users (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  username VARCHAR(64) NOT NULL UNIQUE,
  nickname VARCHAR(64) NOT NULL DEFAULT '',
  password_hash VARCHAR(255) NOT NULL,
  role VARCHAR(20) NOT NULL DEFAULT 'user',
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  INDEX idx_users_role (role)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS categories (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  name VARCHAR(64) NOT NULL UNIQUE,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS site_settings (
  id BIGINT PRIMARY KEY,
  brand_name VARCHAR(120) NOT NULL,
  brand_subtitle VARCHAR(255) NOT NULL,
  brand_icon_path VARCHAR(255) NULL,
  hero_eyebrow VARCHAR(120) NOT NULL,
  hero_title VARCHAR(255) NOT NULL,
  hero_description VARCHAR(500) NOT NULL,
  updated_by_user_id BIGINT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  CONSTRAINT fk_site_settings_updated_by FOREIGN KEY (updated_by_user_id) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS uploads (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  admin_user_id BIGINT NOT NULL,
  original_filename VARCHAR(255) NOT NULL,
  stored_path VARCHAR(255) NOT NULL UNIQUE,
  file_size BIGINT NOT NULL,
  status VARCHAR(20) NOT NULL,
  error_message VARCHAR(512) NOT NULL DEFAULT '',
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT fk_uploads_admin FOREIGN KEY (admin_user_id) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS books (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  title VARCHAR(255) NOT NULL,
  author VARCHAR(255) NOT NULL,
  owner_user_id BIGINT NULL,
  category_id BIGINT NULL,
  description TEXT NOT NULL,
  chapter_count INT NOT NULL DEFAULT 0,
  latest_chapter_title VARCHAR(255) NOT NULL DEFAULT '',
  recommend_score INT NOT NULL DEFAULT 0,
  cover_path VARCHAR(255) NULL,
  source_upload_id BIGINT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  INDEX idx_books_title (title),
  INDEX idx_books_author (author),
  INDEX idx_books_owner_user_id (owner_user_id),
  INDEX idx_books_category (category_id),
  INDEX idx_books_created_at (created_at),
  INDEX idx_books_recommend_score (recommend_score),
  CONSTRAINT fk_books_owner_user FOREIGN KEY (owner_user_id) REFERENCES users(id),
  CONSTRAINT fk_books_category FOREIGN KEY (category_id) REFERENCES categories(id),
  CONSTRAINT fk_books_upload FOREIGN KEY (source_upload_id) REFERENCES uploads(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS chapters (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  book_id BIGINT NOT NULL,
  chapter_index INT NOT NULL,
  title VARCHAR(255) NOT NULL,
  content LONGTEXT NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE KEY uk_chapters_book_index (book_id, chapter_index),
  INDEX idx_chapters_book_index (book_id, chapter_index),
  CONSTRAINT fk_chapters_book FOREIGN KEY (book_id) REFERENCES books(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
`
