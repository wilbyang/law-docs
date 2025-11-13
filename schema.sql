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


create trigger doc_notify
    after insert or update on documents
    for each row
    execute function notify_document_change();
create or replace function notify_document_change()
    returns trigger as $$
declare
    payload text;
    channel text := 'document_changes';
begin
    payload := json_build_object(
        'operation', TG_OP,
        'schema', TG_TABLE_NAME,
        'table', TG_TABLE_NAME,
        'record', to_jsonb(NEW),
        'old_record', to_jsonb(OLD)
    )::text;

    perform pg_notify(channel, payload);
    return NEW;
end;
$$ language plpgsql;


    