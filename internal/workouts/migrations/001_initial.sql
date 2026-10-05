CREATE TABLE IF NOT EXISTS workouts (
  id TEXT PRIMARY KEY,
  activity_type TEXT NOT NULL,
  start_time TIMESTAMP NOT NULL,
  end_time TIMESTAMP NOT NULL,
  duration INTEGER NOT NULL,
  distance REAL NOT NULL,
  route_json TEXT NOT NULL
);
