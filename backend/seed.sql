INSERT INTO trucks (
  unit_number,
  vin,
  year,
  make,
  company,
  ownership,
  plate_number,
  plate_state,
  status,
  status_changed_at,
  status_note,
  samsara_id,
  dot_inspection_expires_at,
  dot_inspection_form_url,
  next_pm_odometer,
  next_oil_change_odometer,
  notes
)
VALUES
  (
    '063',
    '1FUJGLDR0MS123063',
    2018,
    'Freightliner',
    'MS Exp',
    'MS Exp',
    'TX-0638',
    'TX',
    'ENROUTE',
    '2026-05-01',
    'Assigned to long-haul run',
    'SAM-063',
    '2026-07-18',
    'https://example.com/dot/063',
    412500,
    398500,
    'Clean unit, normal service cadence.'
  ),
  (
    '071',
    '1XP4D49X3LD000071',
    2021,
    'Freightliner',
    'MS Exp',
    'MS Exp',
    'OK-0711',
    'OK',
    'ENROUTE',
    '2026-05-07',
    'Running regional loads',
    'SAM-071',
    '2026-06-20',
    'https://example.com/dot/071',
    385000,
    372000,
    'Primary truck for Texas lanes.'
  ),
  (
    '088',
    '1XK4DP0X9MJ000088',
    2020,
    'Kenworth',
    'Flinn',
    'Flinn',
    'AR-0884',
    'AR',
    'SHOP',
    '2026-05-10',
    'Waiting on PM service',
    'SAM-088',
    '2026-05-29',
    'https://example.com/dot/088',
    469200,
    455000,
    'In for inspection and PM service.'
  ),
  (
    '102',
    '1FUJGLDR2LD000102',
    2019,
    'Peterbilt',
    'Owner O',
    'Owner O',
    'NM-1022',
    'NM',
    'STOP',
    '2026-04-28',
    'Driver vacation hold',
    'SAM-102',
    '2026-08-05',
    'https://example.com/dot/102',
    502000,
    491500,
    'Parked unit, ready to reactivate.'
  ),
  (
    '115',
    '1XP5DB9X7ND000115',
    2022,
    'Volvo',
    'MS Exp',
    'MS Exp',
    'TX-1155',
    'TX',
    'UNAVAILABLE',
    '2026-05-03',
    'Out for body work',
    'SAM-115',
    '2026-09-14',
    'https://example.com/dot/115',
    278900,
    265500,
    'Temporarily unavailable after damage repair.'
  ),
  (
    '141',
    '1M1AN4GY5NM000141',
    2021,
    'International',
    'Flinn',
    'Flinn',
    'LA-1417',
    'LA',
    'ENROUTE',
    '2026-05-08',
    'Freshly dispatched',
    'SAM-141',
    '2026-06-30',
    'https://example.com/dot/141',
    331250,
    317500,
    'Solid daily driver with recent service.'
  )
ON CONFLICT (unit_number) DO UPDATE SET
  vin = EXCLUDED.vin,
  year = EXCLUDED.year,
  make = EXCLUDED.make,
  company = EXCLUDED.company,
  ownership = EXCLUDED.ownership,
  plate_number = EXCLUDED.plate_number,
  plate_state = EXCLUDED.plate_state,
  status = EXCLUDED.status,
  status_changed_at = EXCLUDED.status_changed_at,
  status_note = EXCLUDED.status_note,
  samsara_id = EXCLUDED.samsara_id,
  dot_inspection_expires_at = EXCLUDED.dot_inspection_expires_at,
  dot_inspection_form_url = EXCLUDED.dot_inspection_form_url,
  next_pm_odometer = EXCLUDED.next_pm_odometer,
  next_oil_change_odometer = EXCLUDED.next_oil_change_odometer,
  notes = EXCLUDED.notes,
  active = TRUE,
  updated_at = NOW();

-- Tablet assignments mirrored from Fleet board.xlsx for the sample trucks
INSERT INTO tablets (
  truck_id,
  imei,
  phone_number,
  device_make,
  device_model,
  contract_type,
  contract_start,
  contract_end,
  status,
  notes
)
SELECT
  t.id,
  '354154262580035',
  '904-671-4884',
  'Samsung',
  'Galaxy Tab A9+ 5G SM-X218U 64GB Graphite',
  'Installment',
  '2025-04-30',
  '2028-04-30',
  'Active',
  NULL
FROM trucks t
WHERE t.unit_number = '071'
  AND NOT EXISTS (
    SELECT 1
    FROM tablets existing
    WHERE existing.imei = '354154262580035'
  );

INSERT INTO tablets (
  truck_id,
  imei,
  phone_number,
  device_make,
  device_model,
  contract_type,
  contract_start,
  contract_end,
  status,
  notes
)
SELECT
  t.id,
  NULL,
  NULL,
  NULL,
  NULL,
  NULL,
  NULL,
  NULL,
  'Personal',
  'Her own tablet'
FROM trucks t
WHERE t.unit_number = '088'
  AND NOT EXISTS (
    SELECT 1
    FROM tablets existing
    WHERE existing.truck_id = t.id
      AND existing.notes = 'Her own tablet'
  );
