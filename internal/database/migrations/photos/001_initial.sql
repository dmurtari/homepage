CREATE TABLE albums (
    id TEXT NOT NULL UNIQUE PRIMARY KEY,
    name TEXT NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT ''
);

CREATE TABLE photos (
    slug TEXT NOT NULL UNIQUE PRIMARY KEY,
    original TEXT NOT NULL,
    date TEXT NOT NULL DEFAULT '',

    thumb_width INTEGER NOT NULL DEFAULT 0,
    thumb_height INTEGER NOT NULL DEFAULT 0,
    full_width INTEGER NOT NULL DEFAULT 0,
    full_height INTEGER NOT NULL DEFAULT 0,

    description TEXT NOT NULL DEFAULT '',
    location TEXT NOT NULL DEFAULT '',

    camera TEXT NOT NULL DEFAULT '',
    aperture TEXT NOT NULL DEFAULT '',
    iso TEXT NOT NULL DEFAULT '',
    shutter TEXT NOT NULL DEFAULT '',

    gps_lat REAL,
    gps_lon REAL,

    album_id TEXT,
    FOREIGN KEY (album_id) REFERENCES albums(id) ON DELETE SET NULL
);
