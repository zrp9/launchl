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
	name varchar(150) not null,
	details	text not null,
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
	role_id INTEGER NOT NULL REFERENCES roles(id),
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
	name varchar(255) not null,
	active boolean not null default false,
	created_at timestamptz not null default current_timestamp,
	updated_at timestamptz not null default current_timestamp
);

create table if not exists survey_questions (
	id uuid default uuid_generate_v4() primary key,
	survey_id uuid not null references survey (id),
	question_type question_type not null default 'check',
	options jsonb not null,
	active boolean not null default false,
	created_at timestamptz not null default current_timestamp,
	updated_at timestamptz not null default current_timestamp
)

create table if not exists user_surveys (
	user_id UUID NOT NULL,
    survey_id UUID NOT NULL,
    PRIMARY KEY (user_id, survey_id),
    CONSTRAINT fk_user
        FOREIGN KEY (user_id)
        REFERENCES users (id)
        ON DELETE CASCADE,
    CONSTRAINT fk_survey
        FOREIGN KEY (survey_id)
        REFERENCES surveys (id)
        ON DELETE CASCADE
);
