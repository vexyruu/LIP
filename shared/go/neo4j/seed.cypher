// Nodes
MERGE (:User { user_id: 'user-banned',  banned: true,  risk_score: 0.9 });
MERGE (:User { user_id: 'user-fraud-1', banned: false, risk_score: 0.0 });
MERGE (:User { user_id: 'user-fraud-2', banned: false, risk_score: 0.0 });
MERGE (:User { user_id: 'user-clean',   banned: false, risk_score: 0.0 });

MERGE (:Device  { device_fingerprint: 'fp-abc123' });
MERGE (:IPAddress   { ip: '192.168.1.100' });
MERGE (:PaymentMethod   { payment_id: 'payment-001' });

// Relationships
MATCH (u:User { user_id: 'user-banned' }),  (d:Device { device_fingerprint: 'device-001' })
MERGE (u)-[:USED_DEVICE]->(d);

MATCH (u:User { user_id: 'user-fraud-1' }), (d:Device { device_fingerprint: 'device-001' })
MERGE (u)-[:USED_DEVICE]->(d);

MATCH (u:User { user_id: 'user-fraud-1' }), (i:IPAddress { ip: '192.168.1.100' })
MERGE (u)-[:SHARED_IP]->(i);

MATCH (u:User { user_id: 'user-fraud-2' }), (i:IPAddress { ip: '192.168.1.100' })
MERGE (u)-[:SHARED_IP]->(i);