// Nodes
MERGE (:User { user_id: 'user-banned',  banned: true,  risk_score: 0.9 });
MERGE (:User { user_id: 'user-fraud-1', banned: false, risk_score: 0.0 });
MERGE (:User { user_id: 'user-fraud-2', banned: false, risk_score: 0.0 });
MERGE (:User { user_id: 'user-clean',   banned: false, risk_score: 0.0 });

// Postgres sellers (see scripts/seed-dev.sql)
MERGE (:User { user_id: '00000000-0000-0000-0000-000000000002', banned: false, risk_score: 0.0 });
MERGE (:User { user_id: '11111111-1111-1111-1111-111111111111', banned: false, risk_score: 0.0 });

MERGE (:Device { device_fingerprint: 'device-001' });
MERGE (:IPAddress { ip: '192.168.1.100' });
MERGE (:PaymentMethod { payment_id: 'payment-001' });

// Fraud ring (synthetic users)
MATCH (u:User { user_id: 'user-banned' }), (d:Device { device_fingerprint: 'device-001' })
MERGE (u)-[:USED_DEVICE]->(d);

MATCH (u:User { user_id: 'user-fraud-1' }), (d:Device { device_fingerprint: 'device-001' })
MERGE (u)-[:USED_DEVICE]->(d);

MATCH (u:User { user_id: 'user-fraud-1' }), (i:IPAddress { ip: '192.168.1.100' })
MERGE (u)-[:SHARED_IP]->(i);

MATCH (u:User { user_id: 'user-fraud-2' }), (i:IPAddress { ip: '192.168.1.100' })
MERGE (u)-[:SHARED_IP]->(i);

// Fraud demo seller only, shares device + payment with banned user (→ HIGH risk floor 0.9)
MATCH (u:User { user_id: '11111111-1111-1111-1111-111111111111' }), (d:Device { device_fingerprint: 'device-001' })
MERGE (u)-[:USED_DEVICE]->(d);

MATCH (seller:User { user_id: '11111111-1111-1111-1111-111111111111' }), (banned:User { user_id: 'user-banned' })
MERGE (seller)-[:USED_PAYMENT]->(banned);
