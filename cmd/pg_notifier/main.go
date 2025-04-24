package main

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

/*
 *
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
*/
// insert into documents(title, content, doc_size, created_at, updated_at, status, author_id, file_path) select title, content, doc_size, created_at, updated_at, status, author_id, file_path from documents order by id limit 1;
func main() {
	ctx := context.TODO()
	//pgx v5 connection
	pgpool, err := pgxpool.New(ctx, "postgres://boya:@localhost:28813/law_docs")

	if err != nil {
		panic("Unable to connect to database: " + err.Error())
	}
	defer pgpool.Close()
	conn, err := pgpool.Acquire(context.Background())
	if err != nil {
		log.Fatalf("Unable to acquire connection: %v\n", err)
	}
	defer conn.Release()

	_, err = conn.Exec(context.Background(), "LISTEN document_changes")
	if err != nil {
		log.Fatalf("Unable to listen to task_queue: %v\n", err)
	}

	for {
		notification, err := conn.Conn().WaitForNotification(context.Background())
		if err != nil {
			log.Printf("Error waiting for notification: %v\n", err)
			continue
		}

		log.Printf("Received notification: %s\n", notification.Payload)

	}

}
