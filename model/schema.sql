create table if not exists stories (
  id bigint generated always as identity not null primary key,
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
  scraper text not null
);

create table if not exists story_history (
  id bigint generated always as identity not null primary key,
  story_id bigint not null references stories(id) on delete cascade,
  field text not null,
  old_val text not null,
  new_val text not null,
  created_at timestamp not null
);