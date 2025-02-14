create table if not exists stories (
  id bigserial not null primary key,
  ref_id text not null,
  url text not null,
  by text not null,
  published_at timestamp with time zone not null,
  updated_at timestamp with time zone not null default current_timestamp,
  created_at timestamp with time zone not null default current_timestamp,
  last_seen_fp timestamp with time zone not null default current_timestamp,
  title text not null,
  type text not null,
  score integer not null,
  num_comments integer not null,
  scraper text not null,
  deleted boolean not null default false
);

create index if not exists stories_created_at_idx on stories(created_at);
create unique index if not exists stories_ref_id_idx on stories(ref_id);

create table if not exists story_history (
  id bigserial not null primary key,
  story_id bigint not null references stories(id) on delete cascade,
  field text not null,
  old_val text not null,
  new_val text not null,
  created_at timestamp with time zone not null default current_timestamp
);

drop trigger if exists story_changes ON stories;

create or replace function record_story_changes()
  returns trigger 
  language plpgsql
  as
$$
begin
  NEW.updated_at := current_timestamp;

	if NEW.score != OLD.score then
		 insert into story_history(story_id, field, old_val, new_val)
		 values(OLD.id, 'score', OLD.score, NEW.score);
	end if;

	if NEW.num_comments != OLD.num_comments then
		 insert into story_history(story_id, field, old_val, new_val)
		 values(OLD.id, 'num_comments', OLD.num_comments, NEW.num_comments );
	end if;

	IF NEW.url != OLD.url then
		 insert into story_history(story_id, field, old_val, new_val)
		 values(OLD.id, 'url', OLD.url, NEW.url);
	end if;

	if NEW.title != OLD.title then
		 insert into story_history(story_id, field, old_val, new_val)
		 values(OLD.id, 'title', OLD.title, NEW.title);
	end if;

	if NEW.deleted != OLD.deleted then
		 insert into story_history(story_id, field, old_val, new_val)
		 values(OLD.id, 'deleted', OLD.deleted, NEW.deleted);
	end if;

	if NEW.type != OLD.type then
		 insert into story_history(story_id, field, old_val, new_val)
		 values(OLD.id, 'type', OLD.type, NEW.type);
	end if;

	return NEW;
end;
$$
;

create trigger story_changes before update on stories for each row execute procedure record_story_changes();
