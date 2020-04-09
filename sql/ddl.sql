-- Mtr_Version,Start_Time,Status,Host,Hop,Ip,Loss%,Snt, ,Last,Avg,Best,Wrst,StDev,
-- MTR.UNKNOWN,1586138816,OK,nj-us-ping.vultr.com,5,219.158.6.189,0.00,60,0,31.63,28.55,24.39,32.38,2.25

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
    host        varchar(50),
    result      varchar(50)
) DISTRIBUTED BY (host);

drop table if exists final_report;
create table final_report (
    testDate        timestamp,
    hostname        varchar(50),
    Speed           int,
    avg_lossrate    NUMERIC(100,2),
    max_lossrate    NUMERIC(100,2),
    avg_latency     NUMERIC(100,2),
    max_latency     NUMERIC(100,2)
) DISTRIBUTED BY (hostname);


/* -- query -- 
insert into final_report 
    select 
        to_timestamp(testdate, 'YYYY-MM-DD HH24:MI')::timestamp as testDate,
        hostname,
        result::int as Speed,
        avg_lossrate,
        max_lossrate,
        max_latency,
        avg_latency
    from 
        (select host as hostname,
            max(lossrate) as max_lossrate, 
            avg(lossrate)::numeric(100,2) as avg_lossrate, 
            max(worst) as max_latency, 
            max(avg) as avg_latency
            from mtr_report
            where ip != '???' 
            group by host ) m,
        download_report d 
    where d.host = m.hostname 
    order by max_lossrate,Speed;
*/