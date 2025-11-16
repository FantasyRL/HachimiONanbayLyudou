create table users(
    id VARCHAR(32) NOT NULL PRIMARY KEY,
    name VARCHAR(32) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT current_timestamp,
    updated_at TIMESTAMP(6) WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
    deleted_at TIMESTAMP NULL
);

comment on table users is '用户表';
comment on column users.id is '用户ID,使用oauthID';
comment on column users.name is '用户名称';
comment on column users.created_at is '创建时间';
comment on column users.updated_at is '更新时间';
comment on column users.deleted_at is '删除时间';

create table conversations (
    id           uuid        NOT NULL PRIMARY KEY,
    user_id      varchar(32) NOT NULL,
    messages     jsonb       NOT NULL,
    is_summarized smallint   NOT NULL DEFAULT 0,
    title        varchar(128),
    created_at   TIMESTAMP NOT NULL DEFAULT now(),
    updated_at TIMESTAMP(6) WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
    deleted_at   TIMESTAMP
);

create index idx_conversations_user_id
    on conversations (user_id);

create index idx_conversations_unsummarized
    ON conversations (is_summarized)
    WHERE is_summarized = 0;

comment on table conversations is '对话表';
comment on column conversations.id is '对话ID';
comment on column conversations.user_id is '用户ID';
comment on column conversations.messages is '对话消息，JSON格式存储';
comment on column conversations.is_summarized is '是否已生成摘要，0-否，1-是';
comment on column conversations.title is '对话标题';
comment on column conversations.created_at is '创建时间';
comment on column conversations.updated_at is '更新时间';
comment on column conversations.deleted_at is '删除时间';

create table summaries (
    id           uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    conversation_id uuid     NOT NULL,
    summary_text text        NOT NULL,
    created_at   TIMESTAMP   NOT NULL DEFAULT now(),
    updated_at TIMESTAMP(6) WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
    deleted_at   TIMESTAMP
);
create index idx_summaries_conversation_id
    on summaries (conversation_id);
comment on table summaries is '对话摘要表';
comment on column summaries.id is '摘要ID';
comment on column summaries.conversation_id is '对话ID';
comment on column summaries.summary_text is '摘要内容';
comment on column summaries.created_at is '创建时间';
comment on column summaries.updated_at is '更新时间';
comment on column summaries.deleted_at is '删除时间';

create table todolists(
    id           uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      varchar(32) NOT NULL,
    title        varchar(255) NOT NULL,
    content      text        NOT NULL,
    start_time   TIMESTAMP   NOT NULL,
    end_time     TIMESTAMP   NOT NULL,
    is_all_day    smallint    NOT NULL DEFAULT 0,
    status       smallint    NOT NULL DEFAULT 0,
    remind_at    TIMESTAMP,
    tag          varchar(64),
    created_at   TIMESTAMP   NOT NULL DEFAULT now(),
    updated_at TIMESTAMP(6) WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
    deleted_at   TIMESTAMP
);

create index idx_todolists_user_id
    on todolists (user_id);

comment on table todolists is '待办事项表';
comment on column todolists.id is '待办事项ID';
comment on column todolists.user_id is '用户ID';
comment on column todolists.title is '待办事项标题';
comment on column todolists.content is '待办事项内容';
comment on column todolists.start_time is '待办事项开始时间';
comment on column todolists.end_time is '待办事项结束时间';
comment on column todolists.is_all_day is '是否为全天事项，0-否，1-是';
comment on column todolists.status is '待办事项状态，0-未完成，1-已完成';
comment on column todolists.remind_at is '待办事项提醒时间';
comment on column todolists.tag is '待办事项标签';
comment on column todolists.created_at is '创建时间';
comment on column todolists.updated_at is '更新时间';
comment on column todolists.deleted_at is '删除时间';
