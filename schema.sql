create table if not exists users (
    id serial primary key,
    name text not null,
    email text not null,
    password text not null,
    created_at timestamp default current_timestamp,
    updated_at timestamp default current_timestamp
);

create table if not exists documents (
    id serial primary key,
    title text not null,
    content text not null,
    doc_size integer not null,
    created_at timestamp default current_timestamp,
    updated_at timestamp default current_timestamp,
    meta jsonb,
    status text default 'draft' check (status in ('draft', 'pre-processed', 'auditing', 'audited')),
    author_id integer references users(id),
    file_path text
);



    