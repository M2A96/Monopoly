-- Populate PROPERTY table with standard Monopoly properties
INSERT INTO PROPERTY (property_id, name, color_group, price, house_price, hotel_price, rent, rent_with_color_set, rent_with_1_house, rent_with_2_houses, rent_with_3_houses, rent_with_4_houses, rent_with_hotel, mortgage_value) VALUES
-- Brown properties
(gen_random_uuid(), 'Mediterranean Avenue', 'BROWN', 60, 50, 50, 2, 4, 10, 30, 90, 160, 250, 30),
(gen_random_uuid(), 'Baltic Avenue', 'BROWN', 60, 50, 50, 4, 8, 20, 60, 180, 320, 450, 30),

-- Light Blue properties
(gen_random_uuid(), 'Oriental Avenue', 'LIGHT_BLUE', 100, 50, 50, 6, 12, 30, 90, 270, 400, 550, 50),
(gen_random_uuid(), 'Vermont Avenue', 'LIGHT_BLUE', 100, 50, 50, 6, 12, 30, 90, 270, 400, 550, 50),
(gen_random_uuid(), 'Connecticut Avenue', 'LIGHT_BLUE', 120, 50, 50, 8, 16, 40, 100, 300, 450, 600, 60),

-- Pink properties
(gen_random_uuid(), 'St. Charles Place', 'PINK', 140, 100, 100, 10, 20, 50, 150, 450, 625, 750, 70),
(gen_random_uuid(), 'States Avenue', 'PINK', 140, 100, 100, 10, 20, 50, 150, 450, 625, 750, 70),
(gen_random_uuid(), 'Virginia Avenue', 'PINK', 160, 100, 100, 12, 24, 60, 180, 500, 700, 900, 80),

-- Orange properties
(gen_random_uuid(), 'St. James Place', 'ORANGE', 180, 100, 100, 14, 28, 70, 200, 550, 750, 950, 90),
(gen_random_uuid(), 'Tennessee Avenue', 'ORANGE', 180, 100, 100, 14, 28, 70, 200, 550, 750, 950, 90),
(gen_random_uuid(), 'New York Avenue', 'ORANGE', 200, 100, 100, 16, 32, 80, 220, 600, 800, 1000, 100),

-- Red properties
(gen_random_uuid(), 'Kentucky Avenue', 'RED', 220, 150, 150, 18, 36, 90, 250, 700, 875, 1050, 110),
(gen_random_uuid(), 'Indiana Avenue', 'RED', 220, 150, 150, 18, 36, 90, 250, 700, 875, 1050, 110),
(gen_random_uuid(), 'Illinois Avenue', 'RED', 240, 150, 150, 20, 40, 100, 300, 750, 925, 1100, 120),

-- Yellow properties
(gen_random_uuid(), 'Atlantic Avenue', 'YELLOW', 260, 150, 150, 22, 44, 110, 330, 800, 975, 1150, 130),
(gen_random_uuid(), 'Ventnor Avenue', 'YELLOW', 260, 150, 150, 22, 44, 110, 330, 800, 975, 1150, 130),
(gen_random_uuid(), 'Marvin Gardens', 'YELLOW', 280, 150, 150, 24, 48, 120, 360, 850, 1025, 1200, 140),

-- Green properties
(gen_random_uuid(), 'Pacific Avenue', 'GREEN', 300, 200, 200, 26, 52, 130, 390, 900, 1100, 1275, 150),
(gen_random_uuid(), 'North Carolina Avenue', 'GREEN', 300, 200, 200, 26, 52, 130, 390, 900, 1100, 1275, 150),
(gen_random_uuid(), 'Pennsylvania Avenue', 'GREEN', 320, 200, 200, 28, 56, 150, 450, 1000, 1200, 1400, 160),

-- Dark Blue properties
(gen_random_uuid(), 'Park Place', 'DARK_BLUE', 350, 200, 200, 35, 70, 175, 500, 1100, 1300, 1500, 175),
(gen_random_uuid(), 'Boardwalk', 'DARK_BLUE', 400, 200, 200, 50, 100, 200, 600, 1400, 1700, 2000, 200),

-- Railroads (grouped as properties without color group)
(gen_random_uuid(), 'Reading Railroad', NULL, 200, NULL, NULL, 25, 50, 100, 200, NULL, NULL, NULL, 100),
(gen_random_uuid(), 'Pennsylvania Railroad', NULL, 200, NULL, NULL, 25, 50, 100, 200, NULL, NULL, NULL, 100),
(gen_random_uuid(), 'B&O Railroad', NULL, 200, NULL, NULL, 25, 50, 100, 200, NULL, NULL, NULL, 100),
(gen_random_uuid(), 'Short Line', NULL, 200, NULL, NULL, 25, 50, 100, 200, NULL, NULL, NULL, 100),

-- Utilities (grouped as properties without color group)
(gen_random_uuid(), 'Electric Company', NULL, 150, NULL, NULL, 4, 10, NULL, NULL, NULL, NULL, NULL, 75),
(gen_random_uuid(), 'Water Works', NULL, 150, NULL, NULL, 4, 10, NULL, NULL, NULL, NULL, NULL, 75);

-- Create temporary table to store property IDs for reference
CREATE TEMPORARY TABLE property_ids AS
SELECT property_id, name FROM PROPERTY;

-- Populate BOARD_SPACE table with all 40 spaces
INSERT INTO BOARD_SPACE (space_id, position, space_type, name, property_id) VALUES
-- First row (bottom)
(gen_random_uuid(), 0, 'GO', 'GO', NULL),
(gen_random_uuid(), 1, 'PROPERTY', 'Mediterranean Avenue', (SELECT property_id FROM property_ids WHERE name = 'Mediterranean Avenue')),
(gen_random_uuid(), 2, 'COMMUNITY_CHEST', 'Community Chest', NULL),
(gen_random_uuid(), 3, 'PROPERTY', 'Baltic Avenue', (SELECT property_id FROM property_ids WHERE name = 'Baltic Avenue')),
(gen_random_uuid(), 4, 'TAX', 'Income Tax', NULL),
(gen_random_uuid(), 5, 'PROPERTY', 'Reading Railroad', (SELECT property_id FROM property_ids WHERE name = 'Reading Railroad')),
(gen_random_uuid(), 6, 'PROPERTY', 'Oriental Avenue', (SELECT property_id FROM property_ids WHERE name = 'Oriental Avenue')),
(gen_random_uuid(), 7, 'CHANCE', 'Chance', NULL),
(gen_random_uuid(), 8, 'PROPERTY', 'Vermont Avenue', (SELECT property_id FROM property_ids WHERE name = 'Vermont Avenue')),
(gen_random_uuid(), 9, 'PROPERTY', 'Connecticut Avenue', (SELECT property_id FROM property_ids WHERE name = 'Connecticut Avenue')),

-- Second row (left side)
(gen_random_uuid(), 10, 'JAIL', 'Jail / Just Visiting', NULL),
(gen_random_uuid(), 11, 'PROPERTY', 'St. Charles Place', (SELECT property_id FROM property_ids WHERE name = 'St. Charles Place')),
(gen_random_uuid(), 12, 'PROPERTY', 'Electric Company', (SELECT property_id FROM property_ids WHERE name = 'Electric Company')),
(gen_random_uuid(), 13, 'PROPERTY', 'States Avenue', (SELECT property_id FROM property_ids WHERE name = 'States Avenue')),
(gen_random_uuid(), 14, 'PROPERTY', 'Virginia Avenue', (SELECT property_id FROM property_ids WHERE name = 'Virginia Avenue')),
(gen_random_uuid(), 15, 'PROPERTY', 'Pennsylvania Railroad', (SELECT property_id FROM property_ids WHERE name = 'Pennsylvania Railroad')),
(gen_random_uuid(), 16, 'PROPERTY', 'St. James Place', (SELECT property_id FROM property_ids WHERE name = 'St. James Place')),
(gen_random_uuid(), 17, 'COMMUNITY_CHEST', 'Community Chest', NULL),
(gen_random_uuid(), 18, 'PROPERTY', 'Tennessee Avenue', (SELECT property_id FROM property_ids WHERE name = 'Tennessee Avenue')),
(gen_random_uuid(), 19, 'PROPERTY', 'New York Avenue', (SELECT property_id FROM property_ids WHERE name = 'New York Avenue')),

-- Third row (top)
(gen_random_uuid(), 20, 'FREE_PARKING', 'Free Parking', NULL),
(gen_random_uuid(), 21, 'PROPERTY', 'Kentucky Avenue', (SELECT property_id FROM property_ids WHERE name = 'Kentucky Avenue')),
(gen_random_uuid(), 22, 'CHANCE', 'Chance', NULL),
(gen_random_uuid(), 23, 'PROPERTY', 'Indiana Avenue', (SELECT property_id FROM property_ids WHERE name = 'Indiana Avenue')),
(gen_random_uuid(), 24, 'PROPERTY', 'Illinois Avenue', (SELECT property_id FROM property_ids WHERE name = 'Illinois Avenue')),
(gen_random_uuid(), 25, 'PROPERTY', 'B&O Railroad', (SELECT property_id FROM property_ids WHERE name = 'B&O Railroad')),
(gen_random_uuid(), 26, 'PROPERTY', 'Atlantic Avenue', (SELECT property_id FROM property_ids WHERE name = 'Atlantic Avenue')),
(gen_random_uuid(), 27, 'PROPERTY', 'Ventnor Avenue', (SELECT property_id FROM property_ids WHERE name = 'Ventnor Avenue')),
(gen_random_uuid(), 28, 'PROPERTY', 'Water Works', (SELECT property_id FROM property_ids WHERE name = 'Water Works')),
(gen_random_uuid(), 29, 'PROPERTY', 'Marvin Gardens', (SELECT property_id FROM property_ids WHERE name = 'Marvin Gardens')),

-- Fourth row (right side)
(gen_random_uuid(), 30, 'GO_TO_JAIL', 'Go To Jail', NULL),
(gen_random_uuid(), 31, 'PROPERTY', 'Pacific Avenue', (SELECT property_id FROM property_ids WHERE name = 'Pacific Avenue')),
(gen_random_uuid(), 32, 'PROPERTY', 'North Carolina Avenue', (SELECT property_id FROM property_ids WHERE name = 'North Carolina Avenue')),
(gen_random_uuid(), 33, 'COMMUNITY_CHEST', 'Community Chest', NULL),
(gen_random_uuid(), 34, 'PROPERTY', 'Pennsylvania Avenue', (SELECT property_id FROM property_ids WHERE name = 'Pennsylvania Avenue')),
(gen_random_uuid(), 35, 'PROPERTY', 'Short Line', (SELECT property_id FROM property_ids WHERE name = 'Short Line')),
(gen_random_uuid(), 36, 'CHANCE', 'Chance', NULL),
(gen_random_uuid(), 37, 'PROPERTY', 'Park Place', (SELECT property_id FROM property_ids WHERE name = 'Park Place')),
(gen_random_uuid(), 38, 'TAX', 'Luxury Tax', NULL),
(gen_random_uuid(), 39, 'PROPERTY', 'Boardwalk', (SELECT property_id FROM property_ids WHERE name = 'Boardwalk'));

-- Populate CARD table with Chance cards
INSERT INTO CARD (card_id, card_type, description, action_type, action_value) VALUES
(gen_random_uuid(), 'CHANCE', 'Advance to Go (Collect $200)', 'MOVE_TO_POSITION', 0),
(gen_random_uuid(), 'CHANCE', 'Advance to Illinois Ave. If you pass Go, collect $200', 'MOVE_TO_POSITION', 24),
(gen_random_uuid(), 'CHANCE', 'Advance to St. Charles Place. If you pass Go, collect $200', 'MOVE_TO_POSITION', 11),
(gen_random_uuid(), 'CHANCE', 'Advance to nearest Utility. If unowned, you may buy it. If owned, throw dice and pay owner 10 times the amount thrown', 'MOVE_TO_NEAREST_UTILITY', NULL),
(gen_random_uuid(), 'CHANCE', 'Advance to nearest Railroad. If unowned, you may buy it. If owned, pay owner twice the rental', 'MOVE_TO_NEAREST_RAILROAD', NULL),
(gen_random_uuid(), 'CHANCE', 'Bank pays you dividend of $50', 'COLLECT', 50),
(gen_random_uuid(), 'CHANCE', 'Get Out of Jail Free', 'GET_OUT_OF_JAIL_FREE', NULL),
(gen_random_uuid(), 'CHANCE', 'Go Back 3 Spaces', 'MOVE_BACK', 3),
(gen_random_uuid(), 'CHANCE', 'Go to Jail. Go directly to Jail, do not pass Go, do not collect $200', 'GO_TO_JAIL', NULL),
(gen_random_uuid(), 'CHANCE', 'Make general repairs on all your property. For each house pay $25, for each hotel pay $100', 'REPAIRS', NULL),
(gen_random_uuid(), 'CHANCE', 'Pay poor tax of $15', 'PAY', 15),
(gen_random_uuid(), 'CHANCE', 'Take a trip to Reading Railroad. If you pass Go, collect $200', 'MOVE_TO_POSITION', 5),
(gen_random_uuid(), 'CHANCE', 'Take a walk on the Boardwalk. Advance to Boardwalk', 'MOVE_TO_POSITION', 39),
(gen_random_uuid(), 'CHANCE', 'You have been elected Chairman of the Board. Pay each player $50', 'PAY_EACH_PLAYER', 50),
(gen_random_uuid(), 'CHANCE', 'Your building loan matures. Collect $150', 'COLLECT', 150),
(gen_random_uuid(), 'CHANCE', 'You have won a crossword competition. Collect $100', 'COLLECT', 100);

-- Populate CARD table with Community Chest cards
INSERT INTO CARD (card_id, card_type, description, action_type, action_value) VALUES
(gen_random_uuid(), 'COMMUNITY_CHEST', 'Advance to Go (Collect $200)', 'MOVE_TO_POSITION', 0),
(gen_random_uuid(), 'COMMUNITY_CHEST', 'Bank error in your favor. Collect $200', 'COLLECT', 200),
(gen_random_uuid(), 'COMMUNITY_CHEST', 'Doctor''s fee. Pay $50', 'PAY', 50),
(gen_random_uuid(), 'COMMUNITY_CHEST', 'From sale of stock you get $50', 'COLLECT', 50),
(gen_random_uuid(), 'COMMUNITY_CHEST', 'Get Out of Jail Free', 'GET_OUT_OF_JAIL_FREE', NULL),
(gen_random_uuid(), 'COMMUNITY_CHEST', 'Go to Jail. Go directly to Jail, do not pass Go, do not collect $200', 'GO_TO_JAIL', NULL),
(gen_random_uuid(), 'COMMUNITY_CHEST', 'Grand Opera Night. Collect $50 from every player for opening night seats', 'COLLECT_FROM_EACH_PLAYER', 50),
(gen_random_uuid(), 'COMMUNITY_CHEST', 'Holiday fund matures. Receive $100', 'COLLECT', 100),
(gen_random_uuid(), 'COMMUNITY_CHEST', 'Income tax refund. Collect $20', 'COLLECT', 20),
(gen_random_uuid(), 'COMMUNITY_CHEST', 'It is your birthday. Collect $10 from every player', 'COLLECT_FROM_EACH_PLAYER', 10),
(gen_random_uuid(), 'COMMUNITY_CHEST', 'Life insurance matures. Collect $100', 'COLLECT', 100),
(gen_random_uuid(), 'COMMUNITY_CHEST', 'Pay hospital fees of $100', 'PAY', 100),
(gen_random_uuid(), 'COMMUNITY_CHEST', 'Pay school fees of $50', 'PAY', 50),
(gen_random_uuid(), 'COMMUNITY_CHEST', 'Receive $25 consultancy fee', 'COLLECT', 25),
(gen_random_uuid(), 'COMMUNITY_CHEST', 'You are assessed for street repairs. $40 per house, $115 per hotel', 'REPAIRS_ALTERNATIVE', NULL),
(gen_random_uuid(), 'COMMUNITY_CHEST', 'You have won second prize in a beauty contest. Collect $10', 'COLLECT', 10),
(gen_random_uuid(), 'COMMUNITY_CHEST', 'You inherit $100', 'COLLECT', 100);

-- Clean up temporary table
DROP TABLE property_ids;