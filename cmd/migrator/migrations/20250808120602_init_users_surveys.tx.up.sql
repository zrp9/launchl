CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

do $$ begin
	if not exists (select 1 from pg_type where typname = 'role_permission') then
		create type role_permission as enum ('read', 'write', 'read-write', 'all');
	end if;

    if not exists (select 1 from pg_type where typname = 'question_type') then
		create type question_type as enum ('check', 'multi-check', 'drop-down', 'text');
	end if;

end $$;

create table if not exists features(
	id uuid default uuid_generate_v4() primary key,
	title varchar(150) not null,
	name varchar(150) not null,
	details	text not null,
	quick_description text not null,
	created_at timestamptz not null default current_timestamp,
	updated_at timestamptz not null default current_timestamp
);

create table if not exists roles(
	id uuid default uuid_generate_v4() primary key,
	name varchar(255) not null unique,
	permissions role_permission not null default 'read',
	created_at timestamptz not null default current_timestamp,
	updated_at timestamptz not null default current_timestamp
);

create table if not exists users (
	id uuid default uuid_generate_v4() primary key,
	email VARCHAR(150) NOT NULL UNIQUE,
	phone varchar(12) NULL,
	first_name VARCHAR(100) NOT NULL,
	last_name VARCHAR(100) NOT NULL,
	role_id INTEGER NOT NULL REFERENCES roles(id) on delete cascade,
	would_use boolean default false,
	comments text null,
	company_name varchar(150) not null,
	que_position integer not null,
	created_at TIMESTAMPTZ not null DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMPTZ not null DEFAULT CURRENT_TIMESTAMP
)

CREATE INDEX IF NOT EXISTS idx_usr_uid on users (email);
CREATE INDEX IF NOT EXISTS idx_usr_uid on users (would_use);

create table if not exists surveys (
	id uuid default uuid_generate_v4() primary key,
	version varchar(75) not null,
	name varchar(255) not null,
	active boolean not null default false,
	created_at timestamptz not null default current_timestamp,
	updated_at timestamptz not null default current_timestamp
);

create table if not exists survey_questions (
	id uuid default uuid_generate_v4() primary key,
	survey_id uuid not null references survey (id) on delete cascade,
	question_type question_type not null default 'check',
	prompt text not null,
	position integer not null default 0,
	required boolean not null default true,
	metadata jsonb not null default '{}'::jsonb,
	active boolean not null default false,
	created_at timestamptz not null default current_timestamp,
	updated_at timestamptz not null default current_timestamp
);

create table if not exists survey_question_option (
	id uuid default uuid_generate_v4() primary key,
	question_id uuid not null references survey_questions(id) on delete cascade,
	position integer not null default 0,
	label varchar(255) not null,
	value string varchar(255) null
);

create table if not exists survey_responses (
	question_id uuid not null references survey_questions(id) on delete cascade,
	user_id uuid not null references users(id) on delete cascade,
	primary key (question_id, user_id),
	option_id uuid not null references survey_question_option(id),
	written_response text null
);

create table if not exists referals (
	id uuid default uuid_generate_v4() primary key,
	referer_id uuid not null references users (id) on delete cascade,
	referee_id uuid not null references users (id) on delete cascade
);
