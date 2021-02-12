CREATE TABLE IF NOT EXISTS schedule_latest
(
    id          uuid DEFAULT uuid_generate_v4(),
    time        TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

    network     VARCHAR(100)  NOT NULL,
    chain_id    VARCHAR(100)  NOT NULL,
    version     VARCHAR(50)  NOT NULL,
    kind        VARCHAR(100),
    task_id     VARCHAR(100)  NOT NULL,

    latesttime  TIMESTAMP WITH TIME ZONE,
    hash        TEXT,
    height      BIGSERIAL,
    epoch       TEXT,
    nonce       BYTEA,
    retry       BIGINT,

    error       TEXT,

    PRIMARY KEY (id)
);


CREATE INDEX IF NOT EXISTS sch_lst_nvc on schedule_latest(network, chain_id, version,  kind, task_id, time);
