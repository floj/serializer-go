create table if not exists stories (
  -- id integer primary key autoincrement,
  id bigserial not null,
  ref_id text not null,
  url text not null,
  by text not null,
  created_at timestamp not null,
  scraped_at timestamp not null,
  title text not null,
  type text not null,
  score integer not null,
  num_comments integer not null,
  scraper text not null
);