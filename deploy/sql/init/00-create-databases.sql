SELECT 'CREATE DATABASE auth' WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'auth')\gexec
SELECT 'CREATE DATABASE user' WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'user')\gexec
SELECT 'CREATE DATABASE admin' WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'admin')\gexec
SELECT 'CREATE DATABASE common' WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'common')\gexec
