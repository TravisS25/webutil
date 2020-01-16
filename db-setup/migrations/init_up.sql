CREATE TABLE IF NOT EXISTS user_profile(
    id serial primary key,
    email text not null unique,
    first_name text not null,
    last_name text not null
);

CREATE TABLE IF NOT EXISTS user_group(
    id serial primary key,
    user_group text not null
);

CREATE TABLE IF NOT EXISTS database_table(
    id serial primary key,
    name text not null unique,
    display_name text not null unique,
    column_name text not null
);

CREATE TABLE IF NOT EXISTS database_table_user_group_join(
    id serial primary key,
    database_table_id int not null REFERENCES database_table(id) ON UPDATE CASCADE ON DELETE CASCADE,
    user_group_id int not null REFERENCES user_group(id),
    unique(database_table_id, user_group_id)
);

CREATE TABLE IF NOT EXISTS database_action(
    id serial primary key,
    action text not null unique
);

CREATE TABLE IF NOT EXISTS logging(
    id uuid primary key DEFAULT gen_random_uuid(),
    date_created TIMESTAMP default now(),
    data text,
    primary_key_id int,
    been_viewed boolean not null default false,
    database_action_id int REFERENCES database_action(id),
    database_table_id int default 1 REFERENCES database_table(id) ON UPDATE SET DEFAULT ON DELETE SET DEFAULT,
    user_profile_id int REFERENCES user_profile(id)
);

CREATE TABLE IF NOT EXISTS user_session(
    id serial primary key,
    session text not null unique,
    user_bytes text not null,
    url_bytes text not null,
    group_bytes text not null,
    user_profile_id int not null REFERENCES user_profile(id)
);

INSERT INTO database_action(id, action) VALUES (1, 'CREATE') ON CONFLICT DO NOTHING;
INSERT INTO database_action(id, action) VALUES (2, 'UPDATE') ON CONFLICT DO NOTHING;
INSERT INTO database_action(id, action) VALUES (3, 'DELETE') ON CONFLICT DO NOTHING;
