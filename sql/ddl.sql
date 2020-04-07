-- Mtr_Version,Start_Time,Status,Host,Hop,Ip,Loss%,Snt, ,Last,Avg,Best,Wrst,StDev,
-- MTR.UNKNOWN,1586138816,OK,nj-us-ping.vultr.com,5,219.158.6.189,0.00,60,0,31.63,28.55,24.39,32.38,2.25
-- SELECT TIMESTAMP WITH TIME ZONE 'epoch' + 1586138816  * INTERVAL '1 second';

\c thePatriot

drop table if exists test_report;
create table test_report (
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
