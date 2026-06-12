-- Seed data for local development and testing.
-- Password for the user account is: 123456789
-- Run: sqlite3 <db_path> < scripts/seed.sql

-- ─── User ────────────────────────────────────────────────────────────────────
INSERT OR IGNORE INTO user (id, password_hash) VALUES (
  1,
  '$argon2id$v=19$m=65536,t=2,p=1$Vdjo0uWrtHPBACoDykXbQA$LIct89TgPLpUZvYXKUsu1QFlSMY7gHSvYd1FoYcEvYs'
);

-- ─── User settings ───────────────────────────────────────────────────────────
INSERT OR IGNORE INTO user_setting (key, value, updated_at) VALUES
  ('timezone',      'Asia/Saigon',  strftime('%Y-%m-%dT%H:%M:%fZ','now')),
  ('date_format',   'DD/MM/YYYY',   strftime('%Y-%m-%dT%H:%M:%fZ','now'));

-- ─── Self person ─────────────────────────────────────────────────────────────
INSERT OR IGNORE INTO person (id, name, nickname, gender, is_self, other_notes, created_at, updated_at) VALUES
  (1, 'Alex Morgan', 'Alex', 'male', 1, 'This is me.', '2024-01-01T00:00:00.000Z', '2024-01-01T00:00:00.000Z');

-- ─── Labels ──────────────────────────────────────────────────────────────────
INSERT OR IGNORE INTO people_label (id, name, color) VALUES
  (1,  'Close Friend',  '#ef4444'),
  (2,  'Colleague',     '#3b82f6'),
  (3,  'Family',        '#22c55e'),
  (4,  'Acquaintance',  '#a855f7'),
  (5,  'Mentor',        '#f97316'),
  (6,  'Client',        '#06b6d4'),
  (7,  'Neighbor',      '#eab308'),
  (8,  'School Friend', '#ec4899');

-- ─── Relationship types ───────────────────────────────────────────────────────
INSERT OR IGNORE INTO relationship_type (id, name, reverse_name) VALUES
  (1, 'Friend',    'Friend'),
  (2, 'Colleague', 'Colleague'),
  (3, 'Mentor',    'Mentee'),
  (4, 'Mentee',    'Mentor'),
  (5, 'Partner',   'Partner'),
  (6, 'Sibling',   'Sibling'),
  (7, 'Parent',    'Child'),
  (8, 'Child',     'Parent');

UPDATE relationship_type SET inverse_type_id = 4 WHERE id = 3;
UPDATE relationship_type SET inverse_type_id = 3 WHERE id = 4;
UPDATE relationship_type SET inverse_type_id = 8 WHERE id = 7;
UPDATE relationship_type SET inverse_type_id = 7 WHERE id = 8;

-- ─── People ──────────────────────────────────────────────────────────────────
INSERT OR IGNORE INTO person (id, name, nickname, date_of_birth, gender, relationship_type, other_notes, last_contact_at, created_at, updated_at) VALUES
  (2,  'James Carter',    'Jimmy',   '1988-03-14', 'male',   'friend',     'Met at university. Loves hiking.', '2026-05-20T08:30:00.000Z', '2023-06-01T10:00:00.000Z', '2026-05-20T08:30:00.000Z'),
  (3,  'Sophia Nguyen',   'Sophie',  '1992-07-22', 'female', 'colleague',  'Works at DesignHub. Great at Figma.', '2026-06-01T14:00:00.000Z', '2023-09-15T09:00:00.000Z', '2026-06-01T14:00:00.000Z'),
  (4,  'Michael Torres',  'Mike',    '1985-11-05', 'male',   'mentor',     'Senior engineer at BigTech. Gives great career advice.', '2026-04-10T11:00:00.000Z', '2022-03-20T08:00:00.000Z', '2026-04-10T11:00:00.000Z'),
  (5,  'Emily Chen',      'Em',      '1995-02-28', 'female', 'friend',     'Childhood friend. Lives in Hanoi.', '2026-05-15T16:00:00.000Z', '2021-11-10T07:00:00.000Z', '2026-05-15T16:00:00.000Z'),
  (6,  'David Kim',       'Dave',    '1990-09-17', 'male',   'colleague',  'Backend engineer on the same team.', '2026-06-10T09:00:00.000Z', '2024-01-05T08:00:00.000Z', '2026-06-10T09:00:00.000Z'),
  (7,  'Olivia Smith',    'Liv',     '1993-12-03', 'female', 'friend',     'Yoga instructor. Very calming to be around.', '2026-03-28T18:00:00.000Z', '2022-08-01T10:00:00.000Z', '2026-03-28T18:00:00.000Z'),
  (8,  'Lucas Martinez',  'Luca',    '1987-04-20', 'male',   'colleague',  'Product manager. Very detail-oriented.', '2026-06-05T10:30:00.000Z', '2023-02-14T09:00:00.000Z', '2026-06-05T10:30:00.000Z'),
  (9,  'Ava Johnson',     'Ava',     '1998-06-08', 'female', 'friend',     'Photographer. Met at an art exhibition.', '2026-04-22T15:00:00.000Z', '2024-03-01T11:00:00.000Z', '2026-04-22T15:00:00.000Z'),
  (10, 'Noah Williams',   'Noah',    '1991-01-30', 'male',   '',           'Coffee shop owner downtown.', '2026-02-14T10:00:00.000Z', '2023-07-20T08:00:00.000Z', '2026-02-14T10:00:00.000Z'),
  (11, 'Isabella Brown',  'Bella',   '1996-08-14', 'female', 'friend',     'Moved to Singapore last year.', '2026-05-30T20:00:00.000Z', '2022-12-01T09:00:00.000Z', '2026-05-30T20:00:00.000Z'),
  (12, 'Ethan Davis',     'Ethan',   '1983-10-25', 'male',   'mentor',     'Former boss at StartupX. Great networker.', '2026-01-15T11:00:00.000Z', '2021-05-05T08:00:00.000Z', '2026-01-15T11:00:00.000Z'),
  (13, 'Mia Wilson',      'Mia',     '1994-05-11', 'female', 'colleague',  'Data analyst. Always has interesting insights.', '2026-06-08T14:00:00.000Z', '2023-11-01T09:00:00.000Z', '2026-06-08T14:00:00.000Z'),
  (14, 'Liam Anderson',   'Liam',    '1989-03-07', 'male',   'friend',     'Musician. Plays guitar in a local band.', '2026-05-01T21:00:00.000Z', '2022-06-15T10:00:00.000Z', '2026-05-01T21:00:00.000Z'),
  (15, 'Charlotte Lee',   'Charlie', '1997-09-19', 'female', 'colleague',  'Junior developer on the frontend team.', '2026-06-11T09:30:00.000Z', '2025-01-10T08:00:00.000Z', '2026-06-11T09:30:00.000Z');

-- ─── Contact info ─────────────────────────────────────────────────────────────
INSERT OR IGNORE INTO contact_info (id, person_id, type, value, label, position) VALUES
  (1,  2,  'email',    'james.carter@example.com',   'personal', 0),
  (2,  2,  'phone',    '+1-555-0101',                 'mobile',   1),
  (3,  2,  'telegram', '@jimmycarter',                '',         2),
  (4,  3,  'email',    'sophia.nguyen@designhub.io',  'work',     0),
  (5,  3,  'linkedin', 'linkedin.com/in/sophianguyen','',         1),
  (6,  4,  'email',    'mike.torres@bigtech.com',     'work',     0),
  (7,  4,  'phone',    '+1-555-0202',                 'mobile',   1),
  (8,  5,  'email',    'emily.chen@gmail.com',        'personal', 0),
  (9,  5,  'phone',    '+84-900-123456',              'mobile',   1),
  (10, 6,  'email',    'david.kim@company.dev',       'work',     0),
  (11, 6,  'slack',    '@dkim',                       '',         1),
  (12, 7,  'email',    'olivia.smith@yoga.co',        'work',     0),
  (13, 7,  'instagram','@livinyoga',                  '',         1),
  (14, 8,  'email',    'lucas.m@company.dev',         'work',     0),
  (15, 9,  'email',    'ava.j@photocraft.com',        'personal', 0),
  (16, 9,  'instagram','@avajphoto',                  '',         1),
  (17, 10, 'email',    'noah@coffeecorner.sg',        'work',     0),
  (18, 11, 'email',    'bella.brown@gmail.com',       'personal', 0),
  (19, 11, 'telegram', '@bella_sg',                   '',         1),
  (20, 12, 'email',    'ethan.davis@vc.com',          'work',     0),
  (21, 12, 'linkedin', 'linkedin.com/in/ethandavis',  '',         1),
  (22, 13, 'email',    'mia.wilson@company.dev',      'work',     0),
  (23, 14, 'email',    'liam.band@gmail.com',         'personal', 0),
  (24, 15, 'email',    'charlie.lee@company.dev',     'work',     0);

-- ─── Locations ────────────────────────────────────────────────────────────────
INSERT OR IGNORE INTO location (id, person_id, type, city, country, position) VALUES
  (1,  2,  'home', 'New York',    'US', 0),
  (2,  3,  'home', 'Ho Chi Minh', 'VN', 0),
  (3,  3,  'work', 'Ho Chi Minh', 'VN', 1),
  (4,  4,  'work', 'San Francisco','US',0),
  (5,  5,  'home', 'Hanoi',       'VN', 0),
  (6,  6,  'home', 'Ho Chi Minh', 'VN', 0),
  (7,  7,  'home', 'Boston',      'US', 0),
  (8,  8,  'work', 'Ho Chi Minh', 'VN', 0),
  (9,  9,  'home', 'Tokyo',       'JP', 0),
  (10, 10, 'work', 'Singapore',   'SG', 0),
  (11, 11, 'home', 'Singapore',   'SG', 0),
  (12, 12, 'home', 'London',      'GB', 0),
  (13, 13, 'home', 'Ho Chi Minh', 'VN', 0),
  (14, 14, 'home', 'Hanoi',       'VN', 0),
  (15, 15, 'home', 'Ho Chi Minh', 'VN', 0);

-- ─── Label assignments ────────────────────────────────────────────────────────
INSERT OR IGNORE INTO people_label_assignment (person_id, label_id) VALUES
  (2,  1), (2,  8),
  (3,  2),
  (4,  5),
  (5,  1),
  (6,  2),
  (7,  1), (7,  4),
  (8,  2),
  (9,  4),
  (10, 4), (10, 7),
  (11, 1),
  (12, 5),
  (13, 2),
  (14, 4),
  (15, 2);

-- ─── Work history ─────────────────────────────────────────────────────────────
INSERT OR IGNORE INTO work_history (id, person_id, company, title, start_date, end_date, location, description, position) VALUES
  (1,  2,  'TechCorp',     'Software Engineer',        '2012-06', '2016-09', 'New York',     'Full-stack web development.',              0),
  (2,  2,  'StartupY',     'Senior Engineer',           '2016-10', '2020-03', 'Remote',       'Led backend architecture.',                1),
  (3,  2,  'Freelance',    'Consultant',                '2020-04', '',        'Remote',       'Various clients across fintech.',          2),
  (4,  3,  'DesignHub',    'Product Designer',          '2018-01', '',        'Ho Chi Minh',  'UI/UX design for SaaS products.',          0),
  (5,  4,  'BigTech',      'Staff Engineer',            '2015-03', '',        'San Francisco','Platform engineering.',                    0),
  (6,  4,  'StartupX',     'Senior Engineer',           '2010-06', '2015-02', 'New York',     'Helped build the core product.',           1),
  (7,  5,  'VNG Corp',     'Frontend Developer',        '2017-07', '2021-12', 'Ho Chi Minh',  'React web apps.',                          0),
  (8,  5,  'Tiki',         'Frontend Lead',             '2022-01', '',        'Ho Chi Minh',  'Leading frontend team of 6.',              1),
  (9,  6,  'Company Dev',  'Backend Engineer',          '2022-05', '',        'Ho Chi Minh',  'Go microservices.',                        0),
  (10, 12, 'StartupX',     'CTO',                       '2008-01', '2015-06', 'New York',     'Built the engineering team from scratch.', 0),
  (11, 12, 'VC Fund',      'Technical Advisor',         '2015-07', '',        'London',       'Advising portfolio companies.',            1);

-- ─── Important dates ──────────────────────────────────────────────────────────
INSERT OR IGNORE INTO important_date (id, person_id, kind, label, date_value, recurring, notes) VALUES
  (1,  2,  'birthday',     '',                 '1988-03-14', 1, 'Loves craft beer gifts.'),
  (2,  3,  'birthday',     '',                 '1992-07-22', 1, ''),
  (3,  3,  'anniversary',  'Work anniversary', '2018-01-10', 1, 'Started at DesignHub.'),
  (4,  4,  'birthday',     '',                 '1985-11-05', 1, 'Prefers a simple message.'),
  (5,  5,  'birthday',     '',                 '1995-02-28', 1, 'Celebrate together when she visits.'),
  (6,  5,  'other',        'Name day',         '2025-12-05', 0, 'Vietnamese name day tradition.'),
  (7,  6,  'birthday',     '',                 '1990-09-17', 1, ''),
  (8,  7,  'birthday',     '',                 '1993-12-03', 1, 'Send flowers.'),
  (9,  8,  'birthday',     '',                 '1987-04-20', 1, ''),
  (10, 9,  'birthday',     '',                 '1998-06-08', 1, 'She loves matcha cake.'),
  (11, 11, 'birthday',     '',                 '1996-08-14', 1, ''),
  (12, 14, 'birthday',     '',                 '1989-03-07', 1, 'Band performs on his birthday usually.'),
  (13, 15, 'birthday',     '',                 '1997-09-19', 1, '');

-- ─── Reminders ────────────────────────────────────────────────────────────────
INSERT OR IGNORE INTO reminder (id, title, notes, due_date, person_id, completed, recurrence_rule) VALUES
  (1,  'Call James for catch-up',          '',                            '2026-06-20T09:00:00Z', 2,  0, NULL),
  (2,  'Send Sophie design feedback',      'Review the Figma prototype.',  '2026-06-15T10:00:00Z', 3,  0, NULL),
  (3,  'Monthly 1:1 with Mike',            'Ask about promotion path.',   '2026-06-25T14:00:00Z', 4,  0, 'FREQ=MONTHLY'),
  (4,  'Send birthday gift to Emily',      'Order from Tiki before 25th.','2026-07-20T08:00:00Z', 5,  0, NULL),
  (5,  'Team lunch reminder',              'Book the restaurant.',        '2026-06-18T11:00:00Z', NULL,0, NULL),
  (6,  'Follow up with Ethan',             'Re: intro to Series-A fund.', '2026-06-30T10:00:00Z', 12, 0, NULL),
  (7,  'Weekly sync with Charlotte',       '',                            '2026-06-16T09:00:00Z', 15, 0, 'FREQ=WEEKLY'),
  (8,  'Send article to David',            'The Go concurrency article.', '2026-06-14T17:00:00Z', 6,  0, NULL),
  (9,  'Check in with Bella in Singapore', '',                            '2026-07-01T08:00:00Z', 11, 0, NULL),
  (10, 'Plan Ava photo session',           'Anniversary shoot concept.',  '2026-07-10T10:00:00Z', 9,  0, NULL),
  (11, 'Completed: catch-up with Olivia',  '',                            '2026-05-28T18:00:00Z', 7,  1, NULL),
  (12, 'Completed: coffee with Noah',      '',                            '2026-02-14T10:00:00Z', 10, 1, NULL);

UPDATE reminder SET completed_at = '2026-05-28T19:30:00.000Z' WHERE id = 11;
UPDATE reminder SET completed_at = '2026-02-14T10:45:00.000Z' WHERE id = 12;

-- ─── Activities (journal entries) ────────────────────────────────────────────
INSERT OR IGNORE INTO activity (id, title, occurred_at_date, occurred_at_time, content, created_at, updated_at) VALUES
  (1,  'Lunch with James and David',           '2026-05-20', '12:30', 'Had pho at the corner spot. James talked about moving to Berlin. David mentioned a potential project collaboration.', '2026-05-20T13:45:00.000Z', '2026-05-20T13:45:00.000Z'),
  (2,  'Sophie''s design review call',         '2026-06-01', '14:00', 'Reviewed the new dashboard mockups. She has excellent taste. Suggested adding more whitespace to the cards.', '2026-06-01T15:00:00.000Z', '2026-06-01T15:00:00.000Z'),
  (3,  'Coffee with Mike Torres',              '2026-04-10', '10:00', 'Great career conversation. He recommended reading "An Elegant Puzzle" by Will Larson. Shared that his team is hiring.', '2026-04-10T11:30:00.000Z', '2026-04-10T11:30:00.000Z'),
  (4,  'Video call with Emily',                '2026-05-15', '16:00', 'Long overdue catch-up. She is loving her new role at Tiki. We plan to meet when she visits HCMC next month.', '2026-05-15T17:00:00.000Z', '2026-05-15T17:00:00.000Z'),
  (5,  'Team offsite — day 1',                 '2026-06-05', '09:00', 'Full team at Vung Tau. Lucas organized great activities. Charlotte impressed everyone with her presentation.', '2026-06-05T21:00:00.000Z', '2026-06-05T21:00:00.000Z'),
  (6,  'Team offsite — day 2',                 '2026-06-06', '09:00', 'Beach day. Mia brought great playlist. David won the sandcastle contest.', '2026-06-06T20:00:00.000Z', '2026-06-06T20:00:00.000Z'),
  (7,  'Yoga class with Olivia',               '2026-03-28', '18:00', 'First time trying hot yoga — tough but rewarding. Olivia is incredibly patient as an instructor.', '2026-03-28T19:30:00.000Z', '2026-03-28T19:30:00.000Z'),
  (8,  'Ava''s photo exhibition opening',      '2026-04-22', '15:00', 'Beautiful street photography. She captured the chaos of Hanoi brilliantly. Bought a small print.', '2026-04-22T17:00:00.000Z', '2026-04-22T17:00:00.000Z'),
  (9,  'Liam''s band gig at Acoustic Bar',     '2026-05-01', '20:00', 'The band sounded amazing. They played mostly original songs. Met a few of his bandmates.', '2026-05-01T23:00:00.000Z', '2026-05-01T23:00:00.000Z'),
  (10, 'Video call with Bella',                '2026-05-30', '20:00', 'She is settling well in Singapore. New job at a startup seems exciting. Promised to visit next quarter.', '2026-05-30T21:00:00.000Z', '2026-05-30T21:00:00.000Z'),
  (11, 'Intro call with Ethan at VC Fund',     '2026-01-15', '11:00', 'He introduced me to two founders. Very generous with his network. Followed up with a thank-you note.', '2026-01-15T12:00:00.000Z', '2026-01-15T12:00:00.000Z'),
  (12, 'Coffee meeting with Noah',             '2026-02-14', '10:00', 'Talked for 2 hours. He is thinking about opening a second café. Great conversation about community building.', '2026-02-14T12:00:00.000Z', '2026-02-14T12:00:00.000Z'),
  (13, 'Code review session with Charlotte',   '2026-06-11', '09:30', 'Walked through the new auth middleware together. She asked good questions and picks things up fast.', '2026-06-11T10:30:00.000Z', '2026-06-11T10:30:00.000Z'),
  (14, 'Mia data deep-dive',                   '2026-06-08', '14:00', 'She shared retention analysis for Q2. Surprising drop in week-3 cohort. Need to dig deeper with the product team.', '2026-06-08T15:30:00.000Z', '2026-06-08T15:30:00.000Z');

-- Activity ↔ Person links
INSERT OR IGNORE INTO activity_person (activity_id, person_id) VALUES
  (1,  2),  (1,  6),
  (2,  3),
  (3,  4),
  (4,  5),
  (5,  6),  (5,  8),  (5,  13), (5,  15),
  (6,  6),  (6,  8),  (6,  13),
  (7,  7),
  (8,  9),
  (9,  14),
  (10, 11),
  (11, 12),
  (12, 10),
  (13, 15),
  (14, 13);

-- ─── Gifts ────────────────────────────────────────────────────────────────────
INSERT OR IGNORE INTO gift (id, person_id, title, direction, date, notes, amount_cents, currency) VALUES
  (1,  5,  'Birthday flowers',          'given',    '2025-02-28', 'Sent from Hanoi florist.',    8000,  'VND'),
  (2,  2,  'Craft beer sampler box',    'given',    '2025-03-14', 'For his birthday.',           35000, 'USD'),
  (3,  7,  'Yoga mat & blocks',         'given',    '2025-12-25', 'Christmas gift.',             6000,  'USD'),
  (4,  9,  'Printed photo book',        'given',    '2025-12-20', 'Printed from her exhibition.',9000,  'USD'),
  (5,  4,  '"An Elegant Puzzle" book',  'given',    '2025-11-05', 'Thank-you for mentorship.',   2500,  'USD'),
  (6,  3,  'Specialty coffee beans',    'received', '2025-07-22', 'She remembered I like coffee.',0,    'VND'),
  (7,  6,  'Mechanical keyboard',       'received', '2025-09-17', 'For my new home office.',    120000,'VND'),
  (8,  11, 'Merlion souvenir',          'received', '2025-12-01', 'Brought from Singapore.',     0,    'SGD');

-- ─── Journal labels ───────────────────────────────────────────────────────────
INSERT OR IGNORE INTO journal_label (id, name, color) VALUES
  (1, 'Social',      '#3b82f6'),
  (2, 'Work',        '#f97316'),
  (3, 'Mentorship',  '#8b5cf6'),
  (4, 'Personal',    '#22c55e'),
  (5, 'Celebration', '#ec4899'),
  (6, 'Travel',      '#06b6d4'),
  (7, 'Learning',    '#eab308');

-- Journal label assignments (activity_id → label)
INSERT OR IGNORE INTO journal_label_assignment (activity_id, label_id) VALUES
  (1,  1),  -- Lunch with James and David → Social
  (2,  2),  -- Sophie design review → Work
  (3,  3),  (3,  7),  -- Coffee with Mike → Mentorship, Learning
  (4,  1),  (4,  4),  -- Video call with Emily → Social, Personal
  (5,  2),  (5,  1),  -- Team offsite day 1 → Work, Social
  (6,  1),  (6,  2),  -- Team offsite day 2 → Social, Work
  (7,  1),  (7,  4),  -- Yoga with Olivia → Social, Personal
  (8,  1),  (8,  5),  -- Ava's exhibition → Social, Celebration
  (9,  1),  (9,  5),  -- Liam's gig → Social, Celebration
  (10, 1),  (10, 4),  -- Video call with Bella → Social, Personal
  (11, 3),  (11, 2),  -- Ethan intro call → Mentorship, Work
  (12, 1),  (12, 4),  -- Coffee with Noah → Social, Personal
  (13, 2),  (13, 7),  -- Code review with Charlotte → Work, Learning
  (14, 2),  (14, 7);  -- Mia data deep-dive → Work, Learning

-- ─── Relationships ────────────────────────────────────────────────────────────
INSERT OR IGNORE INTO person_relationship (from_person_id, to_person_id, relationship_type_id, notes) VALUES
  (2,  5,  1, 'Met at university, stayed friends since.'),
  (5,  2,  1, ''),
  (3,  8,  2, 'Work on the same product squad.'),
  (8,  3,  2, ''),
  (4,  1,  3, 'Formally mentoring since 2022.'),
  (1,  4,  4, ''),
  (12, 1,  3, 'Career mentor during StartupX days.'),
  (1,  12, 4, ''),
  (2,  14, 1, 'Introduced by a mutual friend.'),
  (14, 2,  1, '');
