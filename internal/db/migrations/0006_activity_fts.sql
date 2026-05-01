CREATE VIRTUAL TABLE activity_fts USING fts5(
  title, content, content='activity', content_rowid='id',
  tokenize='unicode61 remove_diacritics 2'
);

INSERT INTO activity_fts(rowid, title, content)
SELECT id, title, content FROM activity;

CREATE TRIGGER activity_ai AFTER INSERT ON activity BEGIN
  INSERT INTO activity_fts(rowid, title, content) VALUES (new.id, new.title, new.content);
END;
CREATE TRIGGER activity_ad AFTER DELETE ON activity BEGIN
  INSERT INTO activity_fts(activity_fts, rowid, title, content) VALUES ('delete', old.id, old.title, old.content);
END;
CREATE TRIGGER activity_au AFTER UPDATE ON activity BEGIN
  INSERT INTO activity_fts(activity_fts, rowid, title, content) VALUES ('delete', old.id, old.title, old.content);
  INSERT INTO activity_fts(rowid, title, content) VALUES (new.id, new.title, new.content);
END;
