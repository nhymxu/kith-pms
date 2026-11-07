CREATE TABLE reminder_delivery (
  reminder_id      INTEGER NOT NULL REFERENCES reminder(id) ON DELETE CASCADE,
  due_at           TEXT NOT NULL,
  channel_id       INTEGER NOT NULL REFERENCES notification_channel(id) ON DELETE CASCADE,
  status           TEXT NOT NULL DEFAULT 'pending',
  attempts         INTEGER NOT NULL DEFAULT 0,
  next_attempt_at  TIMESTAMP NOT NULL,
  fire_at          TIMESTAMP NOT NULL,
  last_error_class TEXT NOT NULL DEFAULT '',
  last_status      INTEGER NOT NULL DEFAULT 0,
  sent_at          TIMESTAMP NULL,
  created_at       TIMESTAMP NOT NULL,
  PRIMARY KEY (reminder_id, due_at, channel_id)
);
CREATE INDEX idx_reminder_delivery_pending ON reminder_delivery (status, next_attempt_at);
CREATE INDEX idx_reminder_delivery_channel ON reminder_delivery (channel_id);
