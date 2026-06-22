create table refresh_tokens(
  id UUID primary key,
  user_id BIGINT not null references users(id) on delete cascade,
  token_hash text not null,
  expires_at timestamp not null,
  created_at timestamp not null default now(),
  revoked boolean not null default false,
  user_agent text,
  ip_address text
);

create index idx_refresh_user on refresh_tokens(user_id)
