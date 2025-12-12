CREATE TABLE download_task (
                               id TEXT PRIMARY KEY NOT NULL,
                               user_id TEXT,
                               file_id TEXT,
                               file_name TEXT,
                               file_size INTEGER,
                               downloaded_size INTEGER DEFAULT 0,
                               progress INTEGER DEFAULT 0,
                               speed INTEGER DEFAULT 0,
                               type INTEGER NOT NULL,
                               url TEXT,
                               path TEXT,
                               virtual_path TEXT,
                               state INTEGER,
                               error_msg TEXT,
                               target_dir TEXT,
                               support_range BOOLEAN DEFAULT FALSE,
                               create_time DATETIME,
                               update_time DATETIME,
                               finish_time DATETIME
);

CREATE INDEX download_task_user_id_index ON download_task (user_id);