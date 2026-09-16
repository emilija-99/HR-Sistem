-- Public holidays used when computing business days for absence requests.
-- country_id NULL means the holiday applies to every country.
CREATE TABLE IF NOT EXISTS holidays (
    id BIGSERIAL PRIMARY KEY,
    country_id BIGINT REFERENCES countries(country_id),
    holiday_date DATE NOT NULL,
    name VARCHAR(100) NOT NULL,
    UNIQUE(country_id, holiday_date)
);

-- Fixed-date Serbian public holidays for 2026 and 2027. Movable holidays
-- (e.g. Orthodox Easter) are intentionally not included.
INSERT INTO holidays (country_id, holiday_date, name)
SELECT c.country_id, v.d::date, v.n
FROM countries c
CROSS JOIN (VALUES
    ('2026-01-01','Nova godina'),
    ('2026-01-02','Nova godina'),
    ('2026-01-07','Božić'),
    ('2026-02-15','Dan državnosti'),
    ('2026-02-16','Dan državnosti'),
    ('2026-05-01','Praznik rada'),
    ('2026-05-02','Praznik rada'),
    ('2026-11-11','Dan primirja'),
    ('2027-01-01','Nova godina'),
    ('2027-01-02','Nova godina'),
    ('2027-01-07','Božić'),
    ('2027-02-15','Dan državnosti'),
    ('2027-02-16','Dan državnosti'),
    ('2027-05-01','Praznik rada'),
    ('2027-05-02','Praznik rada'),
    ('2027-11-11','Dan primirja')
) AS v(d, n)
WHERE c.iso = 'RS'
ON CONFLICT DO NOTHING;
