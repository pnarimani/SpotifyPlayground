CREATE KEYSPACE IF NOT EXISTS recently_played
   WITH replication = {
     'class': 'NetworkTopologyStrategy',
     'my-datacenter-1': 3
   }
   AND durable_writes = true;
