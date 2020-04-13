\c thePatriot

drop table if exists mtr_report;
create table mtr_report (
    mtrVersion  varchar(50),
    startTime   int,
    status      varchar(20),
    host        varchar(50),
    hop         int,
    ip          varchar(100),    -- hostname or IP
    lossRate    NUMERIC(100,2),
    snt         varchar(10),     -- count of ping
    unknown     varchar(10),
    last        NUMERIC(100,2),
    avg         NUMERIC(100,2),
    best        NUMERIC(100,2),
    worst       NUMERIC(100,2),
    stDev       NUMERIC(100,2)
) DISTRIBUTED BY (host);

drop table if exists download_report;
create table download_report (
    testdate    varchar(50),
    vendor      varchar(50),
    host        varchar(50),
    result      varchar(50)
) DISTRIBUTED BY (host);

drop table if exists final_report;
create table final_report (
    testDate        timestamp,
    vendor          varchar(50),
    hostname        varchar(50),
    Speed           int,
    avg_lossrate    NUMERIC(100,2),
    max_lossrate    NUMERIC(100,2),
    avg_latency     NUMERIC(100,2),
    max_latency     NUMERIC(100,2)
) DISTRIBUTED BY (hostname);


/* -- query -- 

*/