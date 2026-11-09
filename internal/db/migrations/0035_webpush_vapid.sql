CREATE TABLE webpush_vapid (
  id          INTEGER PRIMARY KEY CHECK (id = 1),
  public_key  TEXT NOT NULL,
  private_enc BLOB NOT NULL,
  created_at  TIMESTAMP NOT NULL
);
