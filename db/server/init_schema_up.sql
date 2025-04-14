-- Create GAME table
CREATE TABLE GAME (
    game_id UUID PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    status VARCHAR(20) CHECK (status IN ('WAITING', 'IN_PROGRESS', 'COMPLETED')),
    current_player_id UUID,
    winner_id UUID
);

-- Create PLAYER table
CREATE TABLE PLAYER (
    player_id SERIAL PRIMARY KEY,
    game_id INTEGER REFERENCES GAME(game_id),
    name VARCHAR(255) NOT NULL,
    balance INTEGER DEFAULT 1500,
    position INTEGER DEFAULT 0,
    in_jail BOOLEAN DEFAULT FALSE,
    jail_turns INTEGER DEFAULT 0,
    bankrupt BOOLEAN DEFAULT FALSE
);

-- Add foreign key constraints to GAME table
ALTER TABLE GAME
ADD CONSTRAINT fk_current_player
FOREIGN KEY (current_player_id) REFERENCES PLAYER(player_id),
ADD CONSTRAINT fk_winner
FOREIGN KEY (winner_id) REFERENCES PLAYER(player_id);

-- Create PROPERTY table
CREATE TABLE PROPERTY (
    property_id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    color_group VARCHAR(20) CHECK (color_group IN ('BROWN', 'LIGHT_BLUE', 'PINK', 'ORANGE', 'RED', 'YELLOW', 'GREEN', 'DARK_BLUE')),
    price INTEGER NOT NULL,
    house_price INTEGER,
    hotel_price INTEGER,
    rent INTEGER NOT NULL,
    rent_with_color_set INTEGER,
    rent_with_1_house INTEGER,
    rent_with_2_houses INTEGER,
    rent_with_3_houses INTEGER,
    rent_with_4_houses INTEGER,
    rent_with_hotel INTEGER,
    mortgage_value INTEGER
);

-- Create PLAYER_PROPERTY table
CREATE TABLE PLAYER_PROPERTY (
    player_id INTEGER REFERENCES PLAYER(player_id),
    property_id INTEGER REFERENCES PROPERTY(property_id),
    mortgaged BOOLEAN DEFAULT FALSE,
    houses INTEGER DEFAULT 0,
    has_hotel BOOLEAN DEFAULT FALSE,
    PRIMARY KEY (player_id, property_id)
);

-- Create BOARD_SPACE table
CREATE TABLE BOARD_SPACE (
    space_id SERIAL PRIMARY KEY,
    position INTEGER UNIQUE NOT NULL,
    space_type VARCHAR(20) CHECK (space_type IN ('PROPERTY', 'CHANCE', 'COMMUNITY_CHEST', 'TAX', 'GO', 'JAIL', 'FREE_PARKING', 'GO_TO_JAIL')),
    name VARCHAR(255) NOT NULL,
    property_id INTEGER REFERENCES PROPERTY(property_id)
);

-- Create CARD table
CREATE TABLE CARD (
    card_id SERIAL PRIMARY KEY,
    card_type VARCHAR(20) CHECK (card_type IN ('CHANCE', 'COMMUNITY_CHEST')),
    description TEXT NOT NULL,
    action_type VARCHAR(50) NOT NULL,
    action_value INTEGER
);

-- Create GAME_LOG table
CREATE TABLE GAME_LOG (
    log_id SERIAL PRIMARY KEY,
    game_id INTEGER REFERENCES GAME(game_id),
    player_id INTEGER REFERENCES PLAYER(player_id),
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    action VARCHAR(255) NOT NULL,
    details TEXT
);

-- Create GAME_HOUSE_HOTEL_BANK table
CREATE TABLE GAME_HOUSE_HOTEL_BANK (
    game_id INTEGER REFERENCES GAME(game_id) PRIMARY KEY,
    houses_remaining INTEGER DEFAULT 32,
    hotels_remaining INTEGER DEFAULT 12
);
