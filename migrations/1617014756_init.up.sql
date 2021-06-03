CREATE TABLE access_data (
    id INTEGER PRIMARY KEY,
    github_username TEXT NOT NULL,
    github_token TEXT NOT NULL,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE TABLE project (
    id INTEGER PRIMARY KEY,
    github_proto TEXT NOT NULL,
    folder_name TEXT NOT NULL,
    folder_repository TEXT,
    package_info TEXT,
    list_of_struct TEXT,
    func_list TEXT,
    type INTEGER,
    version TEXT,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE TABLE project_files (
    id INTEGER PRIMARY KEY,
    file TEXT NOT NULL,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);