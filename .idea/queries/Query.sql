CREATE TABLE file_info (
                           id VARCHAR NOT NULL PRIMARY KEY UNIQUE,
                           name VARCHAR NOT NULL,
                           random_name VARCHAR NOT NULL,
                           size INTEGER NOT NULL,
                           mime VARCHAR NOT NULL,
                           virtual_path TEXT NOT NULL,
                           thumbnail_img TEXT,
                           path TEXT,
                           file_hash TEXT NOT NULL,
                           file_enc_hash TEXT,
                           chunk_signature TEXT,
                           first_chunk_hash TEXT,
                           second_chunk_hash TEXT,
                           third_chunk_hash TEXT,
                           has_full_hash BOOLEAN DEFAULT FALSE,
                           is_enc BOOLEAN,
                           is_chunk BOOLEAN NOT NULL,
                           chunk_count INTEGER,
                           enc_path TEXT NOT NULL,
                           created_at DATETIME,
                           updated_at DATETIME
);

-- 创建索引
CREATE INDEX file_info_index_1 ON file_info (name);
CREATE INDEX file_info_index_0 ON file_info (mime);
CREATE INDEX file_info_file_hash_index ON file_info (file_hash);
CREATE INDEX file_info_chunk_signature_index ON file_info (chunk_signature);