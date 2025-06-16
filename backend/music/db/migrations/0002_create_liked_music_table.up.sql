CREATE TABLE IF NOT EXISTS liked_music (
    id UUID NOT NULL PRIMARY KEY,
    track_id UUID NOT NULL,
    user_id UUID NOT NULL,
    CONSTRAINT liked_music_track_user_uidx UNIQUE (track_id, user_id)
);
