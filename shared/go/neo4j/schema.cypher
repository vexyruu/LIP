CREATE CONSTRAINT user_id IF NOT EXISTS FOR (u:User) REQUIRE u.user_id IS UNIQUE;
CREATE CONSTRAINT device_fingerprint IF NOT EXISTS FOR (d:Device) REQUIRE d.device_fingerprint IS UNIQUE;
CREATE CONSTRAINT ip_address IF NOT EXISTS FOR (i:IPAddress) REQUIRE i.ip IS UNIQUE;
CREATE CONSTRAINT payment_id IF NOT EXISTS FOR (p:PaymentMethod) REQUIRE p.payment_id IS UNIQUE;

CREATE INDEX user_banned IF NOT EXISTS FOR (u:User) ON (u.banned);