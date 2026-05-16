CREATE TABLE IF NOT EXISTS recently_played.plays_by_user (
     user_id text,
     bucket_date text,        -- year-week
     played_at timeuuid,
     track_id text,
     played_at_ts timestamp,
     PRIMARY KEY ((user_id, bucket_date), played_at)
   ) WITH CLUSTERING ORDER BY (played_at DESC)
     AND default_time_to_live = 7776000
     AND compaction = {
       'class': 'TimeWindowCompactionStrategy',
       'compaction_window_unit': 'DAYS',
       'compaction_window_size': '1'
     };
