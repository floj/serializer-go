create table if not exists stories (
  id bigserial not null primary key,
  ref_id text not null,
  url text not null,
  by text not null,
  created_at timestamp not null,
  updated_at timestamp not null default current_timestamp,
  scraped_at timestamp not null,
  title text not null,
  type text not null,
  score integer not null,
  num_comments integer not null,
  scraper text not null,
  deleted boolean not null default false
);

create index if not exists stories_created_at_idx on stories(created_at);
create index if not exists stories_ref_id_idx on stories(ref_id);

create table if not exists story_history (
  id bigserial not null primary key,
  story_id bigint not null references stories(id) on delete cascade,
  field text not null,
  old_val text not null,
  new_val text not null,
  created_at timestamp not null
);

